"""
TestCase: UpgradeFluid
DDC Engine: Alluxio
Steps:
1. clone Fluid repo
2. delete Fluid & check ready
3. install old Fluid & check ready
4. create old dataset, alluxioruntime & check ready
5. update Fluid & check ready
6. create new dataset, alluxioruntime & check ready
7. create old data list pod & new data list pod & check ready
8. check if data list succeed
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


def deleteFluid():
    cmd = "helm delete fluid"
    os.system(cmd)
    print("Delete fluid.")
    time.sleep(60)


def cloneFluidRepo():
    cmd = "git clone https://github.com/fluid-cloudnative/fluid.git"
    os.system(cmd)
    print("Clone fluid repo.")
    time.sleep(1)


def checkFluidDeleted():
    api = client.CoreV1Api()
    while True:
        pods = api.list_namespaced_pod("fluid-system").items
        isDeleted = True
        for pod in pods:
            if pod.status.phase in ("Running", "Pending", "Unknown"):
                isDeleted = False
                break
        if isDeleted:
            print("Fluid deleted.")
            return True
        else:
            print("Fluid deleting.")
        time.sleep(1)


def installFluidWithVersion(version):
    install_cmd = "helm install fluid fluid/charts/fluid/{}".format(version)
    os.system(install_cmd)
    print("install fluid-{}".format(version))
    time.sleep(60)


def upgradeFluid():
    upgrade_cmd = "helm upgrade fluid /fluid/charts/fluid/fluid"
    os.system(upgrade_cmd)
    print("Upgrade fluid.")
    time.sleep(60)


def createDatasetAndRuntime(name):
    api = client.CustomObjectsApi()
    my_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {"name": name},
        "spec": {
            "mounts": [{"mountPoint": "https://mirrors.bit.edu.cn/apache/zookeeper/stable/", "name": "hbase"}]
        }
    }
    my_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {"name": name},
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


def checkDatasetBound(name):
    api = client.CustomObjectsApi()
    while True:
        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=name,
            namespace=namespace,
            plural="datasets"
        )
        if "status" in resource:
            if "phase" in resource["status"]:
                print(resource['status']['phase'])
                if resource["status"]["phase"] == "Bound":
                    return True
            else:
                print(resource)
        else:
            print(resource)
        time.sleep(1)


def checkVolumeResourcesReady(name):
    while True:
        try:
            client.CoreV1Api().read_persistent_volume(name=namespace + "-" + name)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue
        try:
            client.CoreV1Api().read_namespaced_persistent_volume_claim(name=name, namespace=namespace)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue
        print("PersistentVolume & PersistentVolumeClaim Ready.")
        break


def cleanUp(dataset_names, pod_names):
    api = client.CoreV1Api()
    # Delete Data Read Pod
    body = client.V1DeleteOptions(propagation_policy='Background')
    for pod_name in pod_names:
        if getPodNameByPrefix(pod_name, namespace) is not None:
            api.delete_namespaced_pod(name=pod_name, namespace=namespace, body=body)
        print("Delete pod:{}".format(pod_name))
        time.sleep(3)

    # Delete Dataset & Alluxioruntime
    custom_api = client.CustomObjectsApi()
    for dataset_name in dataset_names:
        custom_api.delete_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataset_name,
            namespace=namespace,
            plural="datasets"
        )
        time.sleep(5)
        runtimeDelete = False
        while not runtimeDelete:
            print("{}: runtime still exists...".format(dataset_name))
            try:
                custom_api.get_namespaced_custom_object(
                    group="data.fluid.io",
                    version="v1alpha1",
                    name=dataset_name,
                    namespace=namespace,
                    plural="alluxioruntimes"
                )
            except client.exceptions.ApiException as e:
                if e.status == 404:
                    runtimeDelete = True
                    continue
            time.sleep(1)


def checkFluidReady():
    api = client.CoreV1Api()
    while True:
        time.sleep(1)
        pods = api.list_namespaced_pod("fluid-system").items
        pods = [pod.metadata.name for pod in pods if pod.status.phase == "Running"]
        # check fluidapp-controller
        hasAlluxioruntime = False
        for pod in pods:
            if pod.__contains__("fluidapp-controller"):
                hasAlluxioruntime = True
                break
        if hasAlluxioruntime is False:
            print("fluidapp-controller not running.")
            continue
        print("fluidapp-controller running.")

        # check dataset-controller
        hasDataset = False
        for pod in pods:
            if pod.__contains__("dataset-controller"):
                hasDataset = True
                break
        if hasDataset is False:
            print("dataset-controller not running.")
            continue
        print("dataset-controller running.")

        # check csi-nodeplugin
        hasCsi = False
        for pod in pods:
            if pod.__contains__("csi-nodeplugin-fluid"):
                hasCsi = True
                break
        if hasCsi is False:
            print("csi-nodeplugin-fluid not running.")
            continue
        print("csi-nodeplugin-fluid running.")

        # OK
        print("Fluid Ready.")
        return True


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


def createDataListPod(name, pvc_name):
    api = client.CoreV1Api()
    containers = [client.V1Container(
        name="nginx",
        image="nginx",
        # mount_propagation="HostToContainer"
        volume_mounts=[client.V1VolumeMount(mount_path="/data", name="hbase-vol", mount_propagation="HostToContainer")]
    )]
    volumes = [client.V1Volume(
        name="hbase-vol",
        persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(claim_name=pvc_name)
    )]
    spec = client.V1PodSpec(
        containers=containers,
        volumes=volumes
    )
    pod = client.V1Pod(
        api_version="v1",
        kind="Pod",
        metadata=client.V1ObjectMeta(name=name),
        spec=spec
    )
    api.create_namespaced_pod(namespace=namespace, body=pod)
    print("Pod created.")
    time.sleep(1)


def checkDataListSuccess(name) -> bool:
    cmd = "kubectl -n {} exec -it  {} ls /data/hbase".format(namespace, name)
    success = os.system(cmd)
    if success == 0:
        print("{}: Data Read done.".format(name))
        return True
    else:
        print("{}: Data Read Fail.".format(name))
        return False

def main():
    exit_code = 0
    ### load config
    config.load_incluster_config()

    ### Prepare
    cloneFluidRepo()
    deleteFluid()
    checkFluidDeleted()

    ### Install old Fluid
    installFluidWithVersion("v0.8.0")
    checkFluidReady()

    ### Create old dataset & alluxioruntime
    createDatasetAndRuntime("hbase-old")
    checkDatasetBound("hbase-old")
    checkVolumeResourcesReady("hbase-old")

    ### Upgrade Fluid
    upgradeFluid()
    checkFluidReady()

    ### Create new dataset & alluxioruntime
    createDatasetAndRuntime("hbase-new")
    checkDatasetBound("hbase-new")
    checkVolumeResourcesReady("hbase-new")

    ### Test
    createDataListPod("nginx-old", "hbase-old")
    createDataListPod("nginx-new", "hbase-new")
    if checkPodReady("nginx-old", namespace) and checkPodReady("nginx-new", namespace):
        time.sleep(1)
        if checkDataListSuccess("nginx-new") and checkDataListSuccess("nginx-old"):
            exit_code = 0
        else:
            exit_code = 1
    ### Clean up
    cleanUp(("hbase-old", "hbase-new"), ("nginx-old", "nginx-new"))
    return exit_code


if __name__ == '__main__':
    exit(main())