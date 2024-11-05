"""
TestCase: Scale Runtime
DDC Engine: Alluxio
Steps:
1. patch label
2. create dataset and runtime
3. get the the name of pod where the alluxio worker pod is located
4. create app Pods
5. scale runtime and check where the new alluxio worker is located
6. clean up the environment
"""

from kubernetes import client, config
from kubernetes.client.rest import ApiException

import time


def getNodes():
    api = client.CoreV1Api()
    node_list = api.list_node()
    return node_list


def patchNodeLabel(key, value, node):
    api = client.CoreV1Api()

    body = {
        "metadata": {
            "labels": {
                key: value,
            }
        }
    }

    # Patching the node labels
    api_response = api.patch_node(node.metadata.name, body)
    print("node label: %s\t%s" % (node.metadata.name, node.metadata.labels))



def createDatasetAndRuntime(*runtimes):
    api = client.CustomObjectsApi()
    mirror = "https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/"
    for runtime in runtimes:
        my_dataset = {
            "apiVersion": "data.fluid.io/v1alpha1",
            "kind": "Dataset",
            "metadata": {"name": runtime},
            "spec": {
                "mounts": [{"mountPoint": mirror, "name": "hbase"}]
            }
        }
        print(my_dataset)

        my_alluxioruntime = {
            "apiVersion": "data.fluid.io/v1alpha1",
            "kind": "AlluxioRuntime",
            "metadata": {"name": runtime},
            "spec": {
                "replicas": 1,
                "podMetadata": {
                    "labels": {
                        "foo": "bar"
                    }
                },
                "tieredstore": {
                    "levels": [{
                        "mediumtype": "MEM",
                        "path": "/dev/shm",
                        "quota": "2Gi",
                        "high": "0.95",
                        "low": "0.7"
                    }]
                },
                "properties": {
                    "alluxio.user.block.size.bytes.default":"256MB",
                    "alluxio.user.streaming.reader.chunk.size.bytes":"256MB",
                    "alluxio.user.local.reader.chunk.size.bytes":"256MB",
                    "alluxio.worker.network.reader.buffer.size":"256MB",
                    "alluxio.user.streaming.data.timeout":"300sec",
                },
                "fuse": {
                    "global": True,
                    "nodeSelector": {
                        "fuse": "true"
                    },
                    "args": [
                        "fuse",
                        "--fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,max_readahead=0"
                    ]
                }
            }
        }

        api.create_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            namespace="default",
            plural="datasets",
            body=my_dataset,
        )
        print("Created dataset %s." % (runtime))

        api.create_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            namespace="default",
            plural="alluxioruntimes",
            body=my_alluxioruntime
        )
        print("Created runtime %s" % (runtime))



def checkDatasetBound(dataset):
    api = client.CustomObjectsApi()

    while True:
        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataset,
            namespace="default",
            plural="datasets"
        )

        print(resource)

        if "status" in resource:
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Bound":
                    break
        time.sleep(1)
        print(resource)


def createAppPod(nodeName, podName):
    api = client.CoreV1Api()
    containers = [client.V1Container(
        name="nginx",
        image="nginx",
        volume_mounts=[client.V1VolumeMount(mount_path="/data", name="hbasevol")])
    ]
    volumes=[client.V1Volume(
        name="hbasevol",
        persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(claim_name="hbase"))
    ]
    spec = client.V1PodSpec(
        containers=containers,
        node_selector={"kubernetes.io/hostname": nodeName},
        volumes=volumes
    )
    pod = client.V1Pod(
        api_version="v1",
        kind="Pod",
        metadata=client.V1ObjectMeta(name=podName),
        spec=spec
    )
    try:
        api_response = api.create_namespaced_pod("default", pod)
    except ApiException as e:
        print("Exception when calling CoreV1Api->create_namespaced_pod: %s\n" % e)
    print("Created pod %s" % (pod))

def checkPodRunning(podName):
    api = client.CoreV1Api()

    while True:
        try:
            pod = api.read_namespaced_pod_status(name=podName, namespace="default")
        except ApiException as e:
            if e.status == 404:
                continue
        if pod.status.phase == "Running":
            break
        time.sleep(1)


def getNodeByPod(podName):
    api = client.CoreV1Api()
    try:
        pod = api.read_namespaced_pod(name=podName, namespace="default")
        print("")
        return pod.spec.node_name
    except client.exceptions.ApiException as e:
        if e.status == 404:
            return ""


def patchRuntimeReplicas(runtime, number):
    api = client.CustomObjectsApi()

    resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=runtime,
            namespace="default",
            plural="alluxioruntimes"
        )

    resource['spec']['replicas'] = number
    api.patch_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name=runtime,
        namespace="default",
        plural="alluxioruntimes",
        body=resource
    )
    print(resource)


def cleanDatasetAndRuntime(*datasets):
    custom_api = client.CustomObjectsApi()
    for dataset in datasets:
        custom_api.delete_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataset,
            namespace="default",
            plural="datasets"
        )

        runtimeDelete = False
        while not runtimeDelete:
            print("runtime %s still exists..." % (dataset))
            try:
                runtime = custom_api.get_namespaced_custom_object(
                    group="data.fluid.io",
                    version="v1alpha1",
                    name=dataset,
                    namespace="default",
                    plural="alluxioruntimes"
                )
            except client.exceptions.ApiException as e:
                if e.status == 404:
                    runtimeDelete = True
                    continue

            time.sleep(1)

def cleanAppPod(*pods):
    api = client.CoreV1Api()
    for pod in pods:
        api.delete_namespaced_pod(name=pod, namespace="default")
        time.sleep(1)

def cleanAll(nodeList):
    cleanAppPod("nginx1", "nginx2", "nginx3")
    cleanDatasetAndRuntime("hbase")
    for node in nodeList:
        patchNodeLabel("fuse", None, node)

def main():
    config.load_incluster_config()
    # 1. patch label
    nodes = getNodes()
    nodeList = nodes.items
    if len(nodeList) < 3:
        print("Not enough node")
        return 1
    nodeList = nodeList[0:3]
    nodeNameList = []
    for node in nodeList:
        nodeNameList.append(node.metadata.name)
        patchNodeLabel("fuse", "true", node)
    # 2. create dataset and runtime
    createDatasetAndRuntime("hbase")
    checkDatasetBound("hbase")
    # 3. get the the name of pod where the alluxio worker pod is located
    nodeWhereWorker1Located = getNodeByPod("hbase-worker-0")
    if nodeWhereWorker1Located is None or len(nodeWhereWorker1Located) == 0:
        print("Get node error")
        cleanDatasetAndRuntime("hbase")
        for node in nodeList:
            patchNodeLabel("fuse", None, node)
        return 1
    print("worker0 node: " + nodeWhereWorker1Located)
    nodeNameList.remove(nodeWhereWorker1Located)
    print(nodeNameList)
    # 4. create app Pods
    # nodeNameList[0]: 2 Pods use the dataset
    # nodeNameList[1]: 1 Pod uses the dataset
    createAppPod(nodeNameList[0], "nginx1")
    createAppPod(nodeNameList[0], "nginx2")
    createAppPod(nodeNameList[1], "nginx3")
    checkPodRunning("nginx1")
    checkPodRunning("nginx2")
    checkPodRunning("nginx3")
    # 5. scale runtime and check where the new alluxio worker is located
    patchRuntimeReplicas("hbase", 2)
    checkPodRunning("hbase-worker-1")
    node1 = getNodeByPod("hbase-worker-1")
    # node1 should be nodeNameList[0]
    print("node1: " + node1)
    if node1 != nodeNameList[0]:
        print("alluxio worker scheduled error")
        cleanAll(nodeList)
        return 1
    # 6. clean up the environment
    cleanAll(nodeList)
    return 0

if __name__ == '__main__':
    main()
