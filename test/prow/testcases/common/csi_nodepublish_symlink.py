"""
TestCase: Dataset using symlink for NodePublishVolume 
DDC Engine: Alluxio
Steps:
1. create PV and PVC
2. create Dataset and Runtime
3. check if dataset is bound
4. check if persistentVolumeClaim and PV is created
5. check if app pod is running
6. check csi nodePublishMethod
7. submit data read job
8. clean up
"""

import os
import sys

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.step import SimpleStep, StatusCheckStep, currying_fn, dummy_back
from framework.testflow import TestFlow
from kubernetes import client, config

def check_csi_nodePublishMethod():
    api = client.CoreV1Api()
    pods = api.list_namespaced_pod(namespace="fluid-system")
    for pod in pods.items:
        if "csi-nodeplugin" in pod.metadata.name:
            logs = api.read_namespaced_pod_log(name=pod.metadata.name, namespace="fluid-system", container="plugins")
            if "Creating symlink" in logs:
                return True
    
    return False


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "hbase"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs").set_namespaced_name(namespace, name)
    runtime = fluidapi.assemble_runtime("alluxio-webufs").set_namespaced_name(namespace, name)

    csi_symlink_annotation = "[{\"Labels\": {\"fluid.io/node-puhlish-method\": \"symlink\"}, \"selector\": { \"kind\": \"PersistentVolume\"}}]"
    runtime.set_annotation("data.fluid.io/metadataList", csi_symlink_annotation)

    flow = TestFlow("Common - Test CSI NodePublishVolume symlink")

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
            step_name="check if dataset is bound",
            forth_fn=funcs.check_dataset_bound_fn(name, namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if PV & PVC is ready",
            forth_fn=funcs.check_volume_resource_ready_fn(name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data read job",
            forth_fn=funcs.create_pod_fn(dataset_name=name, name="nginx-test"), 
            back_fn=funcs.delete_pod_fn(name="nginx-test")
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check_pod_running",
            forth_fn=funcs.check_pod_running_fn(name="nginx-test")
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check_nodePublishMethod",
            forth_fn=currying_fn(check_csi_nodePublishMethod)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data read job",
            forth_fn=funcs.create_job_fn("time cp -r /data/zookeeper ./", name),
            back_fn=funcs.delete_job_fn()
        )
    )
    flow.append_step(
        StatusCheckStep(
            step_name="check if data read job success",
            forth_fn=funcs.check_job_status_fn()
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)


if __name__ == '__main__':
    main()
