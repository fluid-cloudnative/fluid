"""
TestCase: CSI Plugin Stale Node Patch Verification
DDC Engine: Alluxio
Steps:
1. create Dataset & Runtime
2. check if dataset is bound
3. create app pod
4. check app pod is running
5. add node label
6. delete app pod
7. create app pod again
8. check app pod is running
9. check added label exist
10. clean up
"""
import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

from kubernetes import client, config

from kubernetes.client.rest import ApiException

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, SleepStep, dummy_back, currying_fn


def getNodes():
    api = client.CoreV1Api()
    node_list = api.list_node()
    return node_list


NS = "default"
def createDatasetAndRuntime():
    api = client.CustomObjectsApi()
    my_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {"name": "hbase", "namespace": NS},
        "spec": {
            "mounts": [{"mountPoint": "https://mirrors.bit.edu.cn/apache/spark/",
            "name": "hbase"}]
        }
    }

    my_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {"name": "hbase", "namespace": NS},
        "spec": {
            "replicas": 1,
            "tieredstore": {
                "levels": [{
                    "mediumtype": "MEM",
                    "path": "/dev/shm",
                    "quota": "2Gi",
                    "high": "0.95",
                    "low": "0.7"
                }]
            }
        }
    }

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace=NS,
        plural="datasets",
        body=my_dataset,
    )

    print("Created dataset.")

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace=NS,
        plural="alluxioruntimes",
        body=my_alluxioruntime
    )

    print("Created alluxioruntime.")

def checkDatasetBound():
    api = client.CustomObjectsApi()

    while True:
        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name="hbase",
            namespace=NS,
            plural="datasets"
        )

        if "status" in resource:
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Bound":
                    break
        print("Not bound.")
        time.sleep(1)
        # print(resource)


def createApp(node_name, dataset_name, namespace="default"):
    api = client.CoreV1Api()
    my_app = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {"name": "nginx", "namespace": namespace},
        "spec": {
            "nodeName": node_name,
            "containers": [
                {
                "name": "nginx",
                "image": "nginx",
                "volumeMounts": [{"mountPath": "/data", "name": "hbase-vol"}]
                }
            ],
            "volumes": [{
                "name": "hbase-vol",
                "persistentVolumeClaim": {
                    "claimName": dataset_name
                }
            }]
        }
    }
    api.create_namespaced_pod(NS, my_app)
    print("Create pod.")

# def checkAppRun():
#     api = client.CoreV1Api()
#     while True:
#         resource = api.read_namespaced_pod("nginx", NS)
#         if (resource.status.phase == "Running"):
#             print("App running.")
#             print(resource.spec)
#             return resource.spec.node_name
#         print("App pod is not running.")
#         time.sleep(1)

def check_app_run(namespace="default"):
    api = client.CoreV1Api()
    resource = api.read_namespaced_pod("nginx", namespace)
    if resource.status.phase == "Running":
        print("Nginx App Running")
        return True

    return False

def addLabel(node_name):
    api = client.CoreV1Api()
    resource = api.read_node(node_name)
    resource.metadata.labels['test-stale'] = 'true'
    api.patch_node(node_name, resource)
    print("Add node label.")

def check_label(node_name):
    api = client.CoreV1Api()
    resource = api.read_node(node_name)
    if (resource.metadata.labels['test-stale'] and resource.metadata.labels['test-stale'] == 'true'):
        print("Added label exists.")
        return True
    else:
        print("Added label does not exist.")
        return False

def delete_label(node_name):
    api = client.CoreV1Api()
    resource = api.read_node(node_name)
    resource.metadata.labels['test-stale'] = None
    api.patch_node(node_name, resource)
    print("Deleted node label.")

def deleteApp(namespace="default"):
    api = client.CoreV1Api()
    while True:
        try:
            api.delete_namespaced_pod("nginx", namespace)
            resource = api.read_namespaced_pod("nginx", namespace)
            print("App pod still exists...")
            time.sleep(1)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                print("Delete pod.")
                return

def checkLabel(node_name):
    api = client.CoreV1Api()
    resource = api.read_node(node_name)
    if (resource.metadata.labels['test-stale'] and resource.metadata.labels['test-stale'] == 'true'):
        print("Added label exists.")
        return True
    else:
        print("Added label does not exist.")
        return False

def cleanUp(node_name):
    deleteApp()

    custom_api = client.CustomObjectsApi()
    custom_api.delete_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name="hbase",
        namespace=NS,
        plural="datasets"
    )

    runtimeDelete = False
    while not runtimeDelete:
        print("runtime still exists...")
        try:
            runtime = custom_api.get_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name="hbase",
                namespace=NS,
                plural="alluxioruntimes"
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                runtimeDelete = True
                continue

        time.sleep(1)
    
    api = client.CoreV1Api()
    resource = api.read_node(node_name)
    resource.metadata.labels['test-stale'] = None
    api.patch_node(node_name, resource)

    print("Delete added label.")
    

def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    nodes = getNodes()
    if len(nodes.items) == 0:
        return 1
    node = nodes.items[0]

    name = "stale-info-check"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
        
    runtime = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
        .set_tieredstore(mediumtype="MEM", path="/dev/shm", quota="2Gi")


    flow = TestFlow("Common - Test Stale Node Info after CSI patch")

    flow.append_step(
        SimpleStep(
            step_name="create dataset",
            forth_fn=funcs.create_dataset_fn(dataset.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(runtime.dump(), name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime",
            forth_fn=funcs.create_runtime_fn(runtime.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check dataset bound",
            forth_fn=funcs.check_dataset_bound_fn(name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data read pod",
            forth_fn=currying_fn(createApp, node_name=node.metadata.name, dataset_name=name, namespace=namespace),
            back_fn=currying_fn(deleteApp, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if data read pod running",
            forth_fn=currying_fn(check_app_run, namespace=namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="patch label on node",
            forth_fn=currying_fn(addLabel, node_name=node.metadata.name),
            back_fn=currying_fn(delete_label, node_name=node.metadata.name)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="recreate data read pod[deleting]",
            forth_fn=currying_fn(deleteApp, namespace=namespace),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="recreate data read pod[creating]",
            forth_fn=currying_fn(createApp, node_name=node.metadata.name, dataset_name=name, namespace=namespace),
            back_fn=currying_fn(deleteApp, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if data read app running",
            forth_fn=currying_fn(check_app_run, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if label still exists",
            forth_fn=currying_fn(check_label, node_name=node.metadata.name),
            timeout=20
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)



    
    # createDatasetAndRuntime()
    # checkDatasetBound()
    # createApp()
    # node_name = checkAppRun()
    # addLabel(node_name)
    # deleteApp()
    # createApp()
    # checkAppRun()
    # res = checkLabel(node_name)
    # cleanUp(node_name)
    # print("Has passed? " + str(True))
    # if not res:
    #     exit(-1)
    # return 0


if __name__ == '__main__':
    main()