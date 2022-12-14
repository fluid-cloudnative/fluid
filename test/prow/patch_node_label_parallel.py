"""
TestCase: Patch Node Label Parallel
DDC Engine: Alluxio
Steps:
1. patch label
2. create dataset and runtime demo1 and demo2
3. check label
4. delete dataset demo1 and create dataset demo3
5. check label
6. clean up
"""

from kubernetes import client, config
from kubernetes.client.rest import ApiException

import time



def getAttributeNode():
    api = client.CoreV1Api()
    node_list = api.list_node()
    if len(node_list.items) > 0:
        return node_list.items[0]
    return None

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

def checkLabel(*datasets, node):
    api = client.CoreV1Api()
    try:
        latestNode = api.read_node(node.metadata.name)
        # print(latestNode)
    except ApiException as e:
        print("Exception when calling CoreV1Api->read_node: %s\n" % e)
        return False
    labels = latestNode.metadata.labels

    dataNumKey = "fluid.io/dataset-num"
    alluxioKeyPrefix = "fluid.io/s-alluxio-default-"
    datasetKeyPrefix = "fluid.io/s-default-"
    # check dataset number label
    # print(labels)
    if dataNumKey not in labels or labels[dataNumKey] != str(len(datasets)):
        return False
    # check alluxio label
    for dataset in datasets:
        alluxioKey = alluxioKeyPrefix + dataset
        datasetKey = datasetKeyPrefix + dataset
        if alluxioKey not in labels or labels[alluxioKey] != "true":
            return False
        if datasetKey not in labels or labels[datasetKey] != "true":
            return False
    return True



def createDatasetAndRuntime(*runtimes):
    api = client.CustomObjectsApi()
    for runtime in runtimes:
        my_dataset = {
            "apiVersion": "data.fluid.io/v1alpha1",
            "kind": "Dataset",
            "metadata": {"name": runtime},
            "spec": {
                "mounts": [{"mountPoint": "https://mirrors.bit.edu.cn/apache/spark/", "name": "hbase"}],
                "nodeAffinity": {
                    "required": {
                        "nodeSelectorTerms": [{
                            "matchExpressions": [{
                                "key": "fluid",
                                "operator": "In",
                                "values": ["multi-dataset"]
                            }]
                        }]
                    }
                },
                "placement": "Shared"
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


def main():
    config.load_kube_config()
    # 1. patch label
    nodes = getNodes()
    if len(nodes.items) == 0:
        return 1
    node = nodes.items[0]
    patchNodeLabel("fluid", "multi-dataset", node)
    # 2. create dataset and runtime demo1 and demo2
    createDatasetAndRuntime("demo1", "demo2")
    checkDatasetBound("demo1")
    checkDatasetBound("demo2")
    time.sleep(20)
    # 3. check label
    if not checkLabel("demo1", "demo2", node=node):
        print("[checkabel] label not found")
        return 1
    # 4. delete dataset demo1 and create dataset demo3
    cleanDatasetAndRuntime("demo1")
    createDatasetAndRuntime("demo3")
    checkDatasetBound("demo3")
    time.sleep(20)
    # 5. check label
    if not checkLabel("demo2", "demo3", node=node):
        print("[checkabel] label not found")
        return 1
    # 6. clean all
    cleanDatasetAndRuntime("demo2", "demo3")
    patchNodeLabel("fluid", None, node)
    return 0


if __name__ == '__main__':
    main()