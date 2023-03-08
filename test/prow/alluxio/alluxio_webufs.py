"""
TestCase: Access WebUFS data
DDC Engine: Alluxio
Steps:
1. create Dataset(WebUFS) & Runtime
2. check if dataset is bound
3. check if persistentVolumeClaim & PV is created
4. submit data read job
5. wait until data read job completes
6. clean up
"""
import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, project_root)

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, dummy_back
from framework.exception import TestError

from kubernetes import client, config

def main():
    config.load_kube_config()

    name = "hbase"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs").set_namespaced_name(namespace, name)
    runtime = fluidapi.assemble_runtime("alluxio-webufs").set_namespaced_name(namespace, name)

    flow = TestFlow("Alluxio - Access webufs data")

    flow.append_step(
        SimpleStep(
            step_name="create dataset",
            forth_fn=funcs.createDatasetFn(dataset.dump()),
            back_fn=funcs.deleteDatasetAndRuntimeFn(runtime.dump(), name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime",
            forth_fn=funcs.createRuntimeFn(runtime.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if dataset is bound",
            forth_fn=funcs.checkDatasetBoundFn(name, namespace)
        )
    )
    flow.append_step(
        StatusCheckStep(
            step_name="check if PV & PVC is ready",
            forth_fn=funcs.checkVolumeResourceReadyFn(name, namespace)
        )
    )
    flow.append_step(
        SimpleStep(
            step_name="create data read job",
            forth_fn=funcs.createJobFn("time cp -r /data/hbase ./", name),
            back_fn=funcs.deleteJobFn()
        )
    )
    flow.append_step(
        StatusCheckStep(
            step_name="check if data read job success",
            forth_fn=funcs.checkJobStatusFn()
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)


if __name__ == '__main__':
    main()