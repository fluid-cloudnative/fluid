"""
TestCase: Recover Fuse
DDC Engine: Alluxio
Steps:
1. check if Fuse Recover is Enabled
2. create Dataset(WebUFS) & Runtime
3. check if dataset is bound
4. check if persistentVolumeClaim & PV is created
5. create data list Pod with Injection label
6. wait until Pod running & check if data list succeed
7. delete Alluxio Fuse Pod
8. check Fuse recovered
9. check if data list succeed
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

from kubernetes import client, config

def getPodNameByPrefix(prefix, pod_namespace):
    api = client.CoreV1Api()
    pods = api.list_namespaced_pod(pod_namespace)
    pods_name = [item.metadata.name for item in pods.items]
    for name in pods_name:
        if name.__contains__(prefix):
            pod_name = name
            return pod_name
    return None


def checkCsiRecoverEnabled() -> bool:
    """
    check if csi-nodeplugin-fluid-xxxx pod.spec.containers has args "FuseRecovery=true"
    """
    fluid_namespace = "fluid-system"
    pod_name = "csi-nodeplugin-fluid"
    pod_name = getPodNameByPrefix(pod_name, fluid_namespace)
    if pod_name is None:
        return False
    api = client.CoreV1Api()
    for i in range(10):
        pod = api.read_namespaced_pod(pod_name, fluid_namespace)
        if str(pod.spec.containers).__contains__("FuseRecovery=true"):
            print("CSI recovery enabled")
            return True
        time.sleep(1)
    return False


def checkFuseRecovered(dataset_name, namespace="default"):
    def getFuseRecoveredUids(namespace="default"):
        api = client.CoreV1Api()
        items = api.list_namespaced_event(namespace=namespace).items
        fuseRecoveryUids = set()
        for item in items:
            if item.message is None:
                continue
            if item.message.__contains__("Fuse recover"):
                fuseRecoveryUids.add(item.involved_object.uid)
        return fuseRecoveryUids

    ### get dataset hbase uid
    api = client.CustomObjectsApi()
    dataset = api.get_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace=namespace,
        plural="datasets",
        name=dataset_name)
    uid = dataset['metadata']['uid']
    print("Dataset uid is: {}".format(uid))
    uids = getFuseRecoveredUids(namespace)
        # print("Total uids are: {}".format(uids))
    if uids.__contains__(uid):
        print("Fuse Recovered.")
        return True
    
    return False


def checkDataListSuccess(pod_name, namespace="default") -> bool:
    cmd = "kubectl -n {} exec -it  {} ls /data/zookeeper".format(namespace, pod_name)
    success = os.system(cmd)
    if success == 0:
        print("Data Read done.")
        return True
    else:
        print("Data Read Fail")
        return False


def deleteAlluxioFusePod(dataset_name, namespace="default"):
    def deletePodWithPrefix(prefix, pod_namespace):
        pod_name = getPodNameByPrefix(prefix, pod_namespace)
        api = client.CoreV1Api()
        api.delete_namespaced_pod(pod_name, pod_namespace)
        time.sleep(5)
        print("Delete pod: {}".format(pod_name))

    deletePodWithPrefix("{}-fuse".format(dataset_name), namespace)
    time.sleep(30)


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()


    if checkCsiRecoverEnabled() is False:
        print("FAIL at checkCsiRecoverEnabled(): FUSE Recover feature gate is not enabled")
        return 1
    
    name = "test-fuse-recover"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, name)

    runtime = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
    
    flow = TestFlow("Common - Test FUSE Recover")

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
            step_name="check pod status",
            forth_fn=funcs.check_pod_running_fn(name="nginx-test", namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="touch FUSE mountpoint",
            forth_fn=currying_fn(checkDataListSuccess, pod_name="nginx-test", namespace=namespace),
            timeout=5
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="delete fuse pod",
            forth_fn=currying_fn(deleteAlluxioFusePod, dataset_name=name, namespace=namespace),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if fuse mountpoint is recovered",
            forth_fn=currying_fn(checkFuseRecovered, dataset_name=name, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="touch FUSE mountpoint",
            forth_fn=currying_fn(checkDataListSuccess, pod_name="nginx-test", namespace=namespace),
            timeout=5
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)


if __name__ == '__main__':
    main()
