"""
TestCase: Resources setting for alluxio runtime
DDC Engine: Alluxio
Steps:
1. create Dataset(WebUFS) & Runtime with specified resource
2. check if dataset is bound
3. check if persistentVolumeClaim & PV is created
4. check if alluxio runtime resources are consistent with expected
5. clean up
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


def check_alluxio_runtime_components_resource_fn(name, namespace="default"):
    def check():
        api = client.CoreV1Api()

        response = api.read_namespaced_pod(name="{}-master-0".format(name), namespace=namespace)
        master_resource_check = True
        for container in response.spec.containers:
            if container.name == "master":  # master
                if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
                        container.resources.requests["cpu"] == "1" and container.resources.requests["memory"] == "4Gi":
                    continue
                else:
                    master_resource_check = False
            elif container.name == "jobmaster":  # jobmaster
                if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
                        container.resources.requests["cpu"] == "1500m" and container.resources.requests["memory"] == "4Gi":
                    continue
                else:
                    master_resource_check = False

        if master_resource_check:
            print("Master Resource Check Pass")

        response = api.read_namespaced_pod(name="{}-worker-0".format(name), namespace=namespace)
        worker_resource_check = True
        for container in response.spec.containers:
            if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
                    container.resources.requests["cpu"] == "1" and container.resources.requests["memory"] == "4Gi":
                continue
            else:
                worker_resource_check = False

        if worker_resource_check:
            print("Worker Resource Check Pass")

        return master_resource_check and worker_resource_check

    return check

def set_alluxio_runtime_resource(runtime):
    runtime.resource["spec"]["master"] = {
        "resources": {
            "requests": {
                "cpu": "1000m",
                "memory": "4Gi"
            },
            "limits": {
                "cpu": "2000m",
                "memory": "8Gi"
            }
        }
    }

    runtime.resource["spec"]["jobMaster"] = {
        "resources": {
            "requests": {
                "cpu": "1500m",
                "memory": "4Gi"
            },
            "limits": {
                "cpu": "2000m",
                "memory": "8Gi"
            }
        }
    }

    runtime.resource["spec"]["worker"] = {
        "resources": {
            "requests": {
                "cpu": "1000m",
                "memory": "4Gi"
            },
            "limits": {
                "cpu": "2000m",
                "memory": "8Gi"
            }
        }
    }

    runtime.resource["spec"]["jobWorker"] = {
        "resources": {
            "requests": {
                "cpu": "1000m",
                "memory": "4Gi"
            },
            "limits": {
                "cpu": "2000m",
                "memory": "8Gi"
            }
        }
    }

    runtime.resource["spec"]["fuse"] = {
        "resources": {
            "requests": {
                "cpu": "1000m",
                "memory": "4Gi"
            },
            "limits": {
                "cpu": "2000m",
                "memory": "8Gi"
            }
        }
    }

def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "alluxio-resources"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs").set_namespaced_name(namespace, name)
    runtime = fluidapi.assemble_runtime("alluxio-webufs").set_namespaced_name(namespace, name).set_tieredstore(mediumtype="MEM", path="/dev/shm", quota="4Gi")
    set_alluxio_runtime_resource(runtime)

    flow = TestFlow("Alluxio - Set AlluxioRuntime resources")

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
        StatusCheckStep(
            step_name="check runtime resources",
            forth_fn=check_alluxio_runtime_components_resource_fn(name, namespace)
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)

if __name__ == '__main__':
    main()
