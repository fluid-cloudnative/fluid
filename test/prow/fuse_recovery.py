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
import time

from kubernetes import client, config

namespace = "default"


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
        time.sleep(5)
    return False


def createDatasetAndRuntime():
    api = client.CustomObjectsApi()
    my_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {"name": "hbase"},
        "spec": {
            "mounts": [{"mountPoint": "https://mirrors.bit.edu.cn/apache/hbase/stable/", "name": "hbase"}]
        }
    }

    my_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {"name": "hbase"},
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
        namespace=namespace,
        plural="datasets",
        body=my_dataset,
    )
    print("Created dataset.")
    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace=namespace,
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
            namespace=namespace,
            plural="datasets"
        )
        print(resource)
        if "status" in resource:
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Bound":
                    time.sleep(5)
                    return True
        time.sleep(1)


def checkFuseRecovered():
    ### get dataset hbase uid
    api = client.CustomObjectsApi()
    dataset = api.get_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace=namespace,
        plural="datasets",
        name="hbase")
    uid = dataset['metadata']['uid']
    print("Dataset hbase uid is: {}".format(uid))
    while True:
        uids = getFuseRecoveredUids()
        print("Total uids are: {}".format(uids))
        if uids.__contains__(uid):
            print("Fuse Recovered.")
            return True
        print("Fuse not Recovered.")
        time.sleep(3)


def checkVolumeResourcesReady():
    while True:
        try:
            client.CoreV1Api().read_persistent_volume(name=namespace + "-hbase")
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue
        try:
            client.CoreV1Api().read_namespaced_persistent_volume_claim(name="hbase", namespace=namespace)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue
        print("PersistentVolume & PersistentVolumeClaim Ready.")
        break
    time.sleep(5)


def createDataListPod(name):
    api = client.CoreV1Api()
    containers = [client.V1Container(
        name="nginx",
        image="nginx",
        # mount_propagation="HostToContainer"
        volume_mounts=[client.V1VolumeMount(mount_path="/data", name="hbase-vol")]
    )]
    volumes = [client.V1Volume(
        name="hbase-vol",
        persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(claim_name="hbase")
    )]
    spec = client.V1PodSpec(
        containers=containers,
        volumes=volumes
    )
    pod = client.V1Pod(
        api_version="v1",
        kind="Pod",
        metadata=client.V1ObjectMeta(name=name, labels={"fuse.serverful.fluid.io/inject": "true"}),
        spec=spec
    )
    api.create_namespaced_pod(namespace=namespace, body=pod)
    print("Pod created.")
    time.sleep(5)


def checkDataListSuccess(name) -> bool:
    cmd = "kubectl -n {} exec -it  {} ls /data/hbase".format(namespace, name)
    success = os.system(cmd)
    if success == 0:
        print("Data Read done.")
        return True
    else:
        print("Data Read Fail")
        return False


def deletePod(prefix, pod_namespace):
    pod_name = getPodNameByPrefix(prefix, pod_namespace)
    api = client.CoreV1Api()
    api.delete_namespaced_pod(pod_name, pod_namespace)
    time.sleep(5)
    print("Delete pod: {}".format(pod_name))


def cleanUp(pod_name):
    api = client.CoreV1Api()
    # Delete Data Read Pod
    body = client.V1DeleteOptions(propagation_policy='Background')
    if getPodNameByPrefix(pod_name, namespace) is not None:
        api.delete_namespaced_pod(name=pod_name, namespace=namespace, body=body)
    print("Delete pod:{}".format(pod_name))
    time.sleep(5)

    # Delete Dataset & Alluxioruntime
    custom_api = client.CustomObjectsApi()
    custom_api.delete_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name="hbase",
        namespace=namespace,
        plural="datasets"
    )
    time.sleep(5)
    runtimeDelete = False
    while not runtimeDelete:
        print("runtime still exists...")
        try:
            runtime = custom_api.get_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name="hbase",
                namespace=namespace,
                plural="alluxioruntimes"
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                runtimeDelete = True
                continue

        time.sleep(1)


def checkPodReady(name, pod_namespace) -> bool:
    api = client.CoreV1Api()
    while True:
        name = getPodNameByPrefix(name, pod_namespace)
        if name is None:
            return False
        pod = api.read_namespaced_pod(name, pod_namespace).status
        if pod.phase == "Running":
            print("Pod-{} is {}.".format(name, pod.phase))
            return True
        else:
            print("Pod-{} is {}.".format(name, pod.phase))
            time.sleep(1)


def deleteAlluxioFusePod():
    deletePod("hbase-fuse", namespace)
    print("Delete Fuse Pod:{}".format("hbase-fuse-xxxx"))
    time.sleep(60)


def getFuseRecoveredUids():
    api = client.CoreV1Api()
    items = api.list_namespaced_event(namespace=namespace).items
    fuseRecoveryUids = set()
    for item in items:
        if item.message.__contains__("Fuse recover"):
            fuseRecoveryUids.add(item.involved_object.uid)
    return fuseRecoveryUids


def main():
    exit_code = 0
    ### Load config
    config.load_incluster_config()
    if checkCsiRecoverEnabled() is False:
        return 1

    ### Create dataset & alluxioruntime
    createDatasetAndRuntime()
    checkDatasetBound()
    checkVolumeResourcesReady()

    ### Create Pod with Injection Label
    createDataListPod("nginx")
    if checkPodReady("nginx", namespace):
        time.sleep(5)
        checkDataListSuccess("nginx")

    ### Delete fuse
    deleteAlluxioFusePod()
    if checkPodReady("hbase-fuse", namespace) and checkFuseRecovered():
        time.sleep(5)
        if checkDataListSuccess("nginx"):
            exit_code = 0
        else:
            exit_code = 1
    cleanUp("nginx")
    return exit_code


if __name__ == '__main__':
    exit(main())
