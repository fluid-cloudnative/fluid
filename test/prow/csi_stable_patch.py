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
import time
from kubernetes import client, config

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


def createApp():
    api = client.CoreV1Api()
    my_app = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {"name": "nginx", "namespace": NS},
        "spec": {
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
                    "claimName": "hbase"
                }
            }]
        }
    }
    api.create_namespaced_pod(NS, my_app)
    print("Create pod.")

def checkAppRun():
    api = client.CoreV1Api()
    while True:
        resource = api.read_namespaced_pod("nginx", NS)
        if (resource.status.phase == "Running"):
            print("App running.")
            print(resource.spec)
            return resource.spec.node_name
        print("App pod is not running.")
        time.sleep(1)

def addLabel(node_name):
    api = client.CoreV1Api()
    resource = api.read_node(node_name)
    resource.metadata.labels['test-stale'] = 'true'
    api.patch_node(node_name, resource)
    print("Add node label.")

def deleteApp():
    api = client.CoreV1Api()
    api.delete_namespaced_pod("nginx", NS)
    while True:
        try:
            resource = api.read_namespaced_pod("nginx", NS)
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
    config.load_incluster_config()
    
    createDatasetAndRuntime()
    checkDatasetBound()
    createApp()
    node_name = checkAppRun()
    addLabel(node_name)
    deleteApp()
    createApp()
    checkAppRun()
    res = checkLabel(node_name)
    cleanUp(node_name)
    print("Has passed? " + str(True))
    if not res:
        exit(-1)
    return 0


if __name__ == '__main__':
    main()