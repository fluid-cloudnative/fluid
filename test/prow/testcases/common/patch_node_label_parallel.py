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

import os
import sys

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

from kubernetes import client, config

from kubernetes.client.rest import ApiException

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, SleepStep, dummy_back, currying_fn

from kubernetes import client, config


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

def checkLabel(datasets, node):
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


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    # 1. patch label
    nodes = getNodes()
    if len(nodes.items) == 0:
        return 1
    node = nodes.items[0]

    namespace = "default"

    dataset1 = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, "demo1") \
        .set_node_affinity("fluid", "multi-dataset") \
        .set_placement("Shared")

    runtime1 = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, "demo1") \
        .set_tieredstore(mediumtype="MEM", path="/dev/shm", quota="4Gi")
    
    dataset2 = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, "demo2") \
        .set_node_affinity("fluid", "multi-dataset") \
        .set_placement("Shared")
    runtime2 = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, "demo2") \
        .set_tieredstore(mediumtype="MEM", path="/dev/shm", quota="4Gi")

    dataset3 = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, "demo3") \
        .set_node_affinity("fluid", "multi-dataset") \
        .set_placement("Shared")
    runtime3 = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, "demo3") \
        .set_tieredstore(mediumtype="MEM", path="/dev/shm", quota="4Gi")

    flow = TestFlow("Common - Patch Node Label in Parallel")

    flow.append_step(
        SimpleStep(
            step_name="patch node label",
            forth_fn=currying_fn(patchNodeLabel, key="fluid", value="multi-dataset", node=node),
            back_fn=currying_fn(patchNodeLabel, key="fluid", value=None, node=node)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create dataset demo1",
            forth_fn=funcs.create_dataset_fn(dataset1.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(runtime1.dump(), "demo1", namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime demo1",
            forth_fn=funcs.create_runtime_fn(runtime1.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create dataset demo2",
            forth_fn=funcs.create_dataset_fn(dataset2.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(runtime2.dump(), "demo2", namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime demo2",
            forth_fn=funcs.create_runtime_fn(runtime2.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check dataset demo1 bound",
            forth_fn=funcs.check_dataset_bound_fn("demo1", namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check dataset demo2 bound",
            forth_fn=funcs.check_dataset_bound_fn("demo2", namespace)
        )
    )

    flow.append_step(
        SleepStep(
            sleep_seconds=20,
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check node label",
            forth_fn=currying_fn(checkLabel, datasets=["demo1", "demo2"], node=node)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="clean up dataset demo1",
            forth_fn=funcs.delete_dataset_and_runtime_fn(runtime1.dump(), "demo1", namespace),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create dataset demo3",
            forth_fn=funcs.create_dataset_fn(dataset3.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(runtime3.dump(), "demo3", namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime demo3",
            forth_fn=funcs.create_runtime_fn(runtime3.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check dataset demo3 bound",
            forth_fn=funcs.check_dataset_bound_fn("demo3", namespace)
        )
    )

    flow.append_step(
        SleepStep(
            sleep_seconds=20
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check node label",
            forth_fn=currying_fn(checkLabel, datasets=["demo2", "demo3"], node=node)
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)



if __name__ == '__main__':
    main()