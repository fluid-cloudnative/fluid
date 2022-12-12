"""
TestCase: Access OSS via JindoRuntime
DDC Engine: JindoFS
Prerequisites:
1. Fluid need to enable JindoRuntime.
2. There is a Secret named 'e2e-test-jindoruntime-secret'
Steps:
1. Create Dataset and JindoRuntime
2. Check if dataset is bound
3. Check if PersistentVolumeClaim & PV is created
4. Submit data read job
5. Wait until data read job completes
6. Clean up
"""
from kubernetes import client, config
import time

namespace = 'default'
dataset_name = 'oss-bucket'
runtime_name = 'oss-bucket'
secret_name = 'e2e-test-jindoruntime-secret'
oss_bucket_name = 'fluid-e2e'
oss_bucket_endpoint = 'oss-cn-hongkong-internal.aliyuncs.com'

def create_dataset_and_jindo_runtime():
    api = client.CustomObjectsApi()
    dataset = {
        'apiVersion': 'data.fluid.io/v1alpha1',
        'kind': 'Dataset',
        'metadata': {
            'name': dataset_name
        },
        'spec': {
            'mounts': [
                {
                    'mountPoint': 'oss://' + oss_bucket_name,
                    'options': {
                        'fs.oss.endpoint': oss_bucket_endpoint
                    },
                    'name': 'ossbucket',
                    'encryptOptions': [
                        {
                          'name': 'fs.oss.accessKeyId',
                          'valueFrom': {
                            'secretKeyRef': {
                              'name': secret_name,
                              'key': 'fs.oss.accessKeyId'
                            }
                          }
                        },
                        {
                          'name': 'fs.oss.accessKeySecret',
                          'valueFrom': {
                            'secretKeyRef': {
                              'name': secret_name,
                              'key': 'fs.oss.accessKeySecret'
                            }
                          }
                        }
                    ]
                }
            ]
        }
    }

    jindo_runtime = {
      'apiVersion': 'data.fluid.io/v1alpha1',
      'kind': 'JindoRuntime',
      'metadata': {
        'name': runtime_name
      },
      'spec': {
        'replicas': 1,
        'tieredstore': {
          'levels': [
            {
              'mediumtype': 'MEM',
              'path': '/dev/shm',
              'quota': '10G',
              'high': '0.99',
              'low': '0.98'
            }
          ]
        }
      }
    }

    api.create_namespaced_custom_object(
        group = 'data.fluid.io',
        version = 'v1alpha1',
        namespace = namespace,
        plural = 'datasets',
        body = dataset
    )

    print("Dataset created.")

    api.create_namespaced_custom_object(
        group = 'data.fluid.io',
        version = 'v1alpha1',
        namespace = namespace,
        plural = 'jindoruntimes',
        body = jindo_runtime
    )

    print("JindoRuntime created.")

    return

def check_dataset_is_bound():
    api = client.CustomObjectsApi()
    while True:
        resource = api.get_namespaced_custom_object(
            group = 'data.fluid.io',
            version = 'v1alpha1',
            name = dataset_name,
            namespace = namespace,
            plural = 'datasets'
        )
        print(resource)
        if 'status' in resource:
            if 'phase' in resource['status']:
                if resource['status']['phase'] == 'Bound':
                    print('Dataset status is bound.')
                    break
        time.sleep(1)
    return

def check_pvc_and_pv_is_created():
    while True:
        try:
            client.CoreV1Api().read_persistent_volume(name = namespace + '-' + dataset_name)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue

        try:
            client.CoreV1Api().read_namespaced_persistent_volume_claim(name = dataset_name, namespace = namespace)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue

        print('PersistentVolume & PersistentVolumeClaim Ready.')
        break
    return

def submit_data_read_job():
    api = client.BatchV1Api()

    container = client.V1Container(
        name = 'busybox',
        image = 'busybox',
        command = ['/bin/sh'],
        args = ['-c', 'set -x; time cp -r /data/ossbucket ./'],
        volume_mounts = [
          client.V1VolumeMount(mount_path = '/data', name = 'oss-bucket-vol')
        ]
    )

    template = client.V1PodTemplateSpec(
        metadata = client.V1ObjectMeta(labels = {'app': 'dataread'}),
        spec = client.V1PodSpec(
          restart_policy = 'Never', 
          containers = [container], 
          volumes = [
            client.V1Volume(
              name = 'oss-bucket-vol',
              persistent_volume_claim = client.V1PersistentVolumeClaimVolumeSource(claim_name = dataset_name)
            )
          ]
        )
    )

    spec = client.V1JobSpec(
        template = template,
        backoff_limit = 4
    )

    job = client.V1Job(
        api_version = 'batch/v1',
        kind = 'Job',
        metadata = client.V1ObjectMeta(name = 'fluid-copy-test'),
        spec = spec
    )

    api.create_namespaced_job(namespace = namespace, body = job)
    print('Job created.')
    return

def check_data_read_job_status():
    api = client.BatchV1Api()

    job_completed = False
    while not job_completed:
        response = api.read_namespaced_job_status(
            name = 'fluid-copy-test',
            namespace = namespace
        )

        if response.status.succeeded is not None or \
            response.status.failed is not None:
            job_completed = True

        time.sleep(1)

    print('Data Read Job done.')
    return

def clean_up():
    batch_api = client.BatchV1Api()

    # Delete Data Read Job

    # See https://github.com/kubernetes-client/python/issues/234
    body = client.V1DeleteOptions(propagation_policy = 'Background')
    batch_api.delete_namespaced_job(name = 'fluid-copy-test', namespace = namespace, body = body)


    custom_api = client.CustomObjectsApi()

    custom_api.delete_namespaced_custom_object(
        group = 'data.fluid.io',
        version = 'v1alpha1',
        name = dataset_name,
        namespace = namespace,
        plural = 'datasets'
    )

    runtimeDelete = False
    while not runtimeDelete:
        print('Runtime still exists...')
        try:
            runtime = custom_api.get_namespaced_custom_object(
                group = 'data.fluid.io',
                version = 'v1alpha1',
                name = runtime_name,
                namespace = namespace,
                plural = 'jindoruntimes'
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                runtimeDelete = True
                continue

        time.sleep(1)

    return

def init_k8s_client():
    config.load_incluster_config()

if __name__ == '__main__':
    init_k8s_client()
    create_dataset_and_jindo_runtime()
    check_dataset_is_bound()
    check_pvc_and_pv_is_created()
    submit_data_read_job()
    check_data_read_job_status()
    clean_up()
