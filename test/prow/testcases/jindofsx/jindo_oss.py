"""
TestCase: Access OSS data after cache warmup
DDC Engine: Jindofsx
Steps:
1. create Dataset & Runtime
2. check if dataset is bound
3. check if persistentVolumeClaim & PV is created
4. submit DataLoad CR
5. wait until DataLoad completes
6. check if dataset cached usage equals to ufs total file size (i.e. Fully cached)
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

def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "demo-dataset"
    namespace = "default"

    dataload_name ="demo-dataset-warmup"

    dataset = fluidapi.assemble_dataset("jindo-oss").set_namespaced_name(namespace, name)
    runtime = fluidapi.assemble_runtime("jindo-oss").set_namespaced_name(namespace, name)
    dataload = fluidapi.DataLoad(dataload_name, namespace) \
        .set_target_dataset(name, namespace) \
        .set_load_metadata(True)

    flow = TestFlow("JindoFS - Access OSS data")

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
            step_name="create dataload",
            forth_fn=funcs.create_dataload_fn(dataload.dump()),
            back_fn=dummy_back,    # DataLoad should have ownerReference of Dataset
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if dataload job completes",
            forth_fn=funcs.check_dataload_job_status_fn(dataload_name, namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if the whole dataset is warmed up",
            forth_fn=funcs.check_dataset_cached_percentage_fn(name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data read job",
            forth_fn=funcs.create_job_fn(script="time cp -r /data/ /tmp-data && [[ ! -z \"$(ls -l /tmp-data)\" ]]", dataset_name=name, namespace=namespace),
            back_fn=funcs.delete_job_fn()
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check data read job status",
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
