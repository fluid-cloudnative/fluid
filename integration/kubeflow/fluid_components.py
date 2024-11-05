# Copyright 2024 The Fluid Authors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from kfp import dsl, compiler

# Create a Fluid dataset which contains data in S3.
@dsl.component(packages_to_install=['git+https://github.com/fluid-cloudnative/fluid-client-python.git'])
def create_s3_dataset(dataset_name: str, namespace: str, mount_point: str, mount_s3_endpoint: str, mount_s3_region: str):
    import logging
    import fluid
    from kubernetes import client
    fluid_client = fluid.FluidClient()

    FLUID_GROUP = "data.fluid.io"
    FLUID_VERSION = "v1alpha1"

    # This is an sample which use some pre-defined options.
    # Users can change these code customily
    dataset = fluid.Dataset(
        api_version="%s/%s" % (FLUID_GROUP, FLUID_VERSION),
        kind="Dataset",
        metadata=client.V1ObjectMeta(
            name=dataset_name,
            namespace=namespace
        ),
        spec=fluid.DatasetSpec(
            mounts=[
                fluid.Mount(
                    mount_point=mount_point,
                    name=dataset_name,
                    options={
                        "alluxio.underfs.s3.endpoint": mount_s3_endpoint,
                        "alluxio.underfs.s3.endpoint.region": mount_s3_region,
                        "alluxio.underfs.s3.disable.dns.buckets": "true",
                        "alluxio.underfs.s3.disable.inherit.acl": "false"
                    },
                    encrypt_options=[
                        {
                            "name": "aws.accessKeyId",
                            "valueFrom": {
                              "secretKeyRef": {
                                "name": "s3-secret",
                                "key": "aws.accessKeyId"
                              }
                            }
                        },
                        {
                            "name": "aws.secretKey",
                            "valueFrom": {
                              "secretKeyRef": {
                                "name": "s3-secret",
                                "key": "aws.secretKey"
                              }
                            }
                        }
                    ]
                )
            ]
        )
    )

    fluid_client.create_dataset(dataset)
    
    logging.info(f"Dataset \"{dataset.metadata.namespace}/{dataset.metadata.name}\" created successfully")

# Deploy a simple AlluxioRuntime
@dsl.component(packages_to_install=['git+https://github.com/fluid-cloudnative/fluid-client-python.git'])
def create_alluxio_runtime(dataset_name: str, namespace: str):
    import logging
    from fluid import AlluxioRuntime, AlluxioRuntimeSpec, models, FluidClient
    from kubernetes import client as k8s_client

    fluid_client = FluidClient()

    FLUID_GROUP = "data.fluid.io"
    FLUID_VERSION = "v1alpha1"

    replicas = 1

    # This is the simplest configuration for AlluxioRuntime, you can change the AlluxioRuntime according to your needs
    alluxio_runtime = AlluxioRuntime(
        api_version="%s/%s" % (FLUID_GROUP, FLUID_VERSION),
        kind="AlluxioRuntime",
        metadata=k8s_client.V1ObjectMeta(
            name=dataset_name,
            namespace=namespace
        ),
        spec=AlluxioRuntimeSpec(
            replicas=replicas,
            tieredstore=models.TieredStore([models.Level('0.95', '0.7', 'MEM', '/dev/shm', '2Gi', volume_type=None)])
        )
    )

    fluid_client.create_runtime(alluxio_runtime)


    logging.info(f"Runtime \"{alluxio_runtime.metadata.namespace}/{alluxio_runtime.metadata.name}\" created successfully")

# Preheat the dataset with specific dataset name and namespace
@dsl.component(packages_to_install=['git+https://github.com/fluid-cloudnative/fluid-client-python.git'])
def preheat_dataset(dataset_name: str, namespace: str):
    import logging
    from fluid import DataLoad, DataLoadSpec, FluidClient
    from kubernetes import client as k8s_client
    
    fluid_client = FluidClient()

    FLUID_GROUP = "data.fluid.io"
    FLUID_VERSION = "v1alpha1"

    dataload = DataLoad(
        api_version="%s/%s" % (FLUID_GROUP, FLUID_VERSION),
        kind="DataLoad",
        metadata=k8s_client.V1ObjectMeta(
            name="%s-loader" % dataset_name,
            namespace=namespace
        ),
        spec=DataLoadSpec(
            dataset={
                "name": dataset_name,
                "namespace": namespace
            }
        )
    )
    
    fluid_client.create_data_operation(data_op=dataload, wait=True)
    
    logging.info(f"Load Dataset \"{namespace}/{dataset_name}\"  successfully")

# Cleanup the dataset along with the corresponding alluxioruntime
@dsl.component(packages_to_install=['git+https://github.com/fluid-cloudnative/fluid-client-python.git'])
def cleanup_dataset_and_alluxio_runtime(dataset_name: str, namespace: str):
    import logging
    from fluid import FluidClient
    
    fluid_client = FluidClient()
    fluid_client.delete_runtime(name=dataset_name, namespace=namespace, runtime_type="alluxio", wait_until_cleaned_up=True)
    fluid_client.delete_dataset(name=dataset_name, namespace=namespace, wait_until_cleaned_up=True)

    logging.info(f"Cleanup Dataset and AlluxioRuntime \"{namespace}/{dataset_name}\" successfully!")

# Cleanup the preheat dataset operation
@dsl.component(packages_to_install=['git+https://github.com/fluid-cloudnative/fluid-client-python.git'])
def cleanup_preheat_operation(dataset_name: str, namespace: str):
    import logging
    from fluid import FluidClient
    
    fluid_client = FluidClient()
    fluid_client.delete_data_operation(name="%s-loader" % dataset_name, data_op_type="dataload", namespace=namespace)
    logging.info("Cleanup preheat dataset operation successfully!")

# Re-run this file when you changed code above to re-generate components' yaml file.
compiler.Compiler().compile(create_s3_dataset, "./component-yaml/create-s3-dataset.yaml")
compiler.Compiler().compile(create_alluxio_runtime, "./component-yaml/create-alluxioruntime.yaml")
compiler.Compiler().compile(preheat_dataset, "./component-yaml/preheat-dataset.yaml")
compiler.Compiler().compile(cleanup_dataset_and_alluxio_runtime, "./component-yaml/cleanup-dataset-and-alluxioruntime.yaml")
compiler.Compiler().compile(cleanup_preheat_operation, "./component-yaml/cleanup-preheat-operation.yaml")