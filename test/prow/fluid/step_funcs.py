import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, project_root)

from kubernetes import client, config
from framework.step import check
from framework.exception import TestError


SERVERLESS_KEY="serverless.fluid.io/inject"
SERVERFUL_KEY="fuse.serverful.fluid.io/inject"
FLUID_MANAGER_KEY="fluid.io/managed-by"

def create_dataset_fn(dataset):
    def create_dataset():

        dataset_namespace = dataset["metadata"]["namespace"]
        api = client.CustomObjectsApi()
        api.create_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            namespace=dataset_namespace,
            plural="datasets",
            body=dataset,
        )
        print("Dataset \"{}/{}\" created".format(dataset["metadata"]["namespace"], dataset["metadata"]["name"]))

    return create_dataset

def check_dataset_bound_fn(name, namespace="default"):
    def check():
        api = client.CustomObjectsApi()

        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=name,
            namespace=namespace,
            plural="datasets"
        )

        if "status" in resource:
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Bound":
                    return True

        return False

    return check

def create_runtime_fn(runtime):
    def create_runtime():
        plural_str = "{}s".format(runtime["kind"].lower())
        runtime_namespace = runtime["metadata"]["namespace"]

        api = client.CustomObjectsApi()
        api.create_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            namespace=runtime_namespace,
            plural=plural_str,
            body=runtime
        )
        print("{} \"{}/{}\" created".format(runtime["kind"], runtime["metadata"]["namespace"],
                                            runtime["metadata"]["name"]))

    return create_runtime

def delete_dataset_and_runtime_fn(runtime, name, namespace="default"):
    def check_clean_up():
        api = client.CustomObjectsApi()
        plural_str = "{}s".format(runtime["kind"].lower())

        print("runtime still exists...")
        try:
            to_delete = api.get_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name=name,
                namespace=namespace,
                plural=plural_str
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                return True

        return False

    def delete_dataset():
        api = client.CustomObjectsApi()

        try:
            api.delete_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name=name,
                namespace=namespace,
                plural="datasets"
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                return
            else:
                raise e
        except Exception as e:
            raise e
            

        print("Dataset \"{}/{}\" deleted".format(namespace, name))

        timeout_check_fn = check(check_clean_up, 60, 3)
        timeout_check_fn()

    return delete_dataset

def create_dataload_fn(dataload):
    def create():
        api = client.CustomObjectsApi()

        name = dataload["metadata"]["name"]
        namespace = dataload["metadata"]["namespace"]

        api.create_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            namespace=namespace,
            plural="dataloads",
            body=dataload,
        )

        print("DataLoad \"{}/{}\" created.".format(namespace, name))

    return create

def check_dataload_job_status_fn(dataload_name, dataload_namespace="default"):
    def check():
        api = client.CustomObjectsApi()

        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataload_name,
            namespace=dataload_namespace,
            plural="dataloads"
        )

        if "status" in resource:
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Complete":
                    return True

        return False

    return check

def check_dataset_cached_percentage_fn(name, namespace="default"):
    def check():
        api = client.CustomObjectsApi()

        resource = api.get_namespaced_custom_object_status(
            group="data.fluid.io",
            version="v1alpha1",
            name=name,
            namespace=namespace,
            plural="datasets"
        )

        if "status" in resource:
            if "ufsTotal" in resource["status"]:
                if "cacheStates" in resource["status"] and "cached" in resource["status"]["cacheStates"]:
                    if resource["status"]["ufsTotal"] == resource["status"]["cacheStates"]["cached"]:
                        print("Checking Dataset warmed up status. Expected: %s, current status: %s" % (
                            resource["status"]["ufsTotal"], resource["status"]["cacheStates"]["cached"]))
                        return True

        return False

    return check

def check_volume_resource_ready_fn(name, namespace="default"):
    def check():
        api = client.CoreV1Api()

        pv_name = "{}-{}".format(namespace, name)
        pvc_name = name

        try:
            api.read_persistent_volume(name=pv_name)
            api.read_namespaced_persistent_volume_claim(name=pvc_name, namespace=namespace)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                return False
        except Exception as e:
            return False

        print("PersistentVolume {} & PersistentVolumeClaim {}/{} ready.".format(pv_name, namespace, pvc_name))
        return True

    return check

def create_job_fn(script, dataset_name, name="fluid-e2e-job-test", namespace="default", serverless=False):
    def create():
        api = client.BatchV1Api()

        container = client.V1Container(
            name="alpine",
            image="alpine",
            command=["/bin/sh"],
            args=["-c", "set -ex; {}".format(script)],
            volume_mounts=[client.V1VolumeMount(mount_path="/data", name="data-vol")]
        )

        if serverless:
            obj_meta=client.V1ObjectMeta(labels={"app": "dataread", SERVERLESS_KEY: "true", FLUID_MANAGER_KEY: "fluid"})
        else:
            obj_meta=client.V1ObjectMeta(labels={"app": "dataread"})

        template = client.V1PodTemplateSpec(
            metadata=obj_meta,
            spec=client.V1PodSpec(restart_policy="Never", containers=[container],
                                  volumes=[client.V1Volume(name="data-vol",
                                                           persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(
                                                               claim_name=dataset_name))])
        )

        spec = client.V1JobSpec(
            template=template,
            backoff_limit=4
        )

        job = client.V1Job(
            api_version="batch/v1",
            kind="Job",
            metadata=client.V1ObjectMeta(name=name),
            spec=spec
        )

        api.create_namespaced_job(namespace=namespace, body=job)
        print("Job \"{}/{}\" created.[script=\"{}\"]".format(namespace, name, script))

    return create

def check_job_status_fn(name="fluid-e2e-job-test", namespace="default"):
    def check():
        api = client.BatchV1Api()

        try:
            response = api.read_namespaced_job_status(
                name=name,
                namespace=namespace
            )

            if response.status.succeeded is not None:
                return True
        except Exception as e:
            print(e)

        return False

    return check

def delete_job_fn(name="fluid-e2e-job-test", namespace="default"):
    def delete():
        batch_api = client.BatchV1Api()

        body = client.V1DeleteOptions(propagation_policy='Background')
        batch_api.delete_namespaced_job(name=name, namespace=namespace, body=body)

    return delete

def create_pod_fn(dataset_name, name="nginx-test", namespace="default", serverless=False, serverful=False):
    def create():
        api = client.CoreV1Api()
        container = client.V1Container(
            name="nginx",
            image="nginx",
            volume_mounts=[client.V1VolumeMount(mount_path="/data", name="data-vol")]
        )

        volume = client.V1Volume(
            name="data-vol",
            persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(claim_name=dataset_name)
        )

        labels = {}
        if serverless:
            labels[SERVERLESS_KEY] = "true"
            labels[FLUID_MANAGER_KEY] = "fluid"
        if serverful:
            labels[SERVERFUL_KEY] = "true"

        pod = client.V1Pod(
            api_version="v1",
            kind="Pod",
            metadata=client.V1ObjectMeta(name=name, labels=labels),
            spec=client.V1PodSpec(
                containers=[container],
                volumes=[volume]
            )
        )

        api.create_namespaced_pod(namespace=namespace, body=pod)
        print("Pod {} created".format(name))

    return create

def check_pod_running_fn(name="nginx-test", namespace="default"):
    def check():
        api = client.CoreV1Api()
        pod_status = api.read_namespaced_pod(name, namespace).status
        if pod_status.phase == "Running":
            return True
        
        return False
    
    return check

def delete_pod_fn(name="nginx-test", namespace="default"):
    def delete():
        api = client.CoreV1Api()
        body = client.V1DeleteOptions(propagation_policy='Background')

        try:
            api.delete_namespaced_pod(name=name, namespace=namespace, body=body)
        except client.exceptions.ApiException as e:
            if e.status != 404:
                raise TestError("failed to delete pod with code status {}".format(e.status))

    return delete
