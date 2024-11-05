"""
TestCase: Webhook mutating pod
DDC Engine: Alluxio
Steps:
1. check if Fuse Recover is Enabled
2. create Dataset(WebUFS) & Runtime
3. check if dataset is bound
4. check if persistentVolumeClaim & PV is created
5. create data list Pod with fuse serverful Injection label
6. wait until Pod running & check if has affinity injection
7. clean up
"""
import os
import sys

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, dummy_back

from kubernetes import client, config


def check_pod_affinity_fn(name, namespace):
    def check():
        api = client.CoreV1Api()
        affinity = api.read_namespaced_pod(name, namespace).spec.affinity
        if len(affinity.node_affinity.preferred_during_scheduling_ignored_during_execution) != 0:
            return True
        return False

    return check


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "test-webhook-mutate"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, name)

    runtime = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
    
    flow = TestFlow("Common - Test Pod Affinity Mutation")

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
        StatusCheckStep(
            step_name="check if PV & PVC is ready",
            forth_fn=funcs.check_volume_resource_ready_fn(name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create pod with fluid pvc",
            forth_fn=funcs.create_pod_fn(dataset_name=name, name="nginx-test", namespace=namespace, serverful=True),
            back_fn=funcs.delete_pod_fn(name="nginx-test", namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check pod affinity",
            forth_fn=check_pod_affinity_fn(name="nginx-test", namespace=namespace)
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)


if __name__ == '__main__':
    main()
