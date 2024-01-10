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

from kfp import dsl, kubernetes, compiler
from kfp import client as kfp_client


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

# compiler.Compiler().compile(create_alluxio_runtime, 'pipeline.yaml')