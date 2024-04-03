## Use vineyard runtime to accelerate kubeflow pipelines

Vineyard can accelerate data sharing by utilizing shared memory compared to existing methods such as local files or S3 services. In this doc, we will show you how to use vineyard runtime to accelerate an existing kubeflow pipeline on the fluid platform.

### Prerequisites

- Install the argo CLI tool via the [official guide](https://github.com/argoproj/argo-workflows/releases/).


### Overview of the pipeline

The pipeline we use is a simple pipeline that trains a linear regression model on the dummy Boston Housing Dataset. It contains three steps: [preprocess](./preprocess-data/preprocess-data.y), [train](./train-data/train-data.py), and [test](./test-data/test-data.py).


### Prepare the environment

- Prepare a kubernetes cluster. If you don't have a kubernetes cluster on hand, you can use the following command to create a kubernetes cluster via kind(v0.20.0+):

```shell
cat <<EOF | kind create cluster -v 5 --name=kind --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.23.0
- role: worker
  image: kindest/node:v1.23.0
- role: worker
  image: kindest/node:v1.23.0
- role: worker
  image: kindest/node:v1.23.0
EOF
```

- Prepare the docker images of the kubeflow pipeline. You can use the following command to build the docker image:

```shell
$ make docker-build REGISTRY="test-registry"(Replace with your custom registry)
```

Then you can push these images to the registry that your kubernetes cluster can access or load these images to the cluster if your kubernetes cluster is created by kind.

```shell
$ make load-images REGISTRY="test-registry"(Replace with your custom registry)
```

or

```shell
$ make push-images REGISTRY="test-registry"(Replace with your custom registry)
```

### Install the fluid platform and kubeflow components

- Refer to the [Installation Documentation](../../../userguide/install.md) to complete the installation.

- Install the argo workflow controller, which can be used as the backend of kubeflow pipeline. You can use the following command to install the argo workflow controller:

```shell
$ kubectl create ns argo
$ kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.4.8/install.yaml
```

- Install the vineyard runtime and dataset. You can use the following command to install the vineyard runtime and dataset:

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard
spec:
  replicas: 2
  tieredstore:
    levels:
    - mediumtype: MEM
      quota: 20Gi
---
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: vineyard
EOF
```

### Run the pipeline

Please make sure the `pipeline.yaml` and `pipeline-with-vineyard.yaml` files are the original 
ones as they have been modified to use the privileged 
container to clear the page cache. Then you can follow
the next steps to run the pipeline:

**Notice** You need to mount a **NAS path** to the kubernetes node.
Here we mount a NAS path to the `/mnt/csi-benchmark`(shown in the `prepare-data.yaml`) path of all kubernetes nodes.
Next, we need to prepare the dataset by running the following command:

```shell
$ kubectl apply -f prepare-data.yaml
```

The dataset will be stored in the host path. Also, you may need to wait for a while for the dataset to be generated and you can use the following command to check the status:

```shell
$ while ! kubectl logs -l app=prepare-data | grep "preparing data time" >/dev/null; do echo "dataset unready, waiting..."; sleep 5; done && echo "dataset ready"
```

Before running the pipeline, we need to create the role and rolebinding for the pipeline as follows.

```shell
$ kubectl apply -f rbac.yaml
```

After that, you can run the pipeline via the following command:

Without vineyard:

```shell
$ argo submit pipeline.yaml -p data_mu
ltiplier=2000 -p registry="test-registry" 
Name:                machine-learning-pipeline-z72gm
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Pending
Created:             Wed Apr 03 11:46:43 +0800 (now)
Progress:            
Parameters:          
  data_multiplier:   2000
  registry:          test-registry
```

The result is shown as follows:

```shell
$ argo get machine-learning-pipeline-z72gm                                           
Name:                machine-learning-pipeline-z72gm
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Succeeded
Conditions:          
 PodRunning          False
 Completed           True
Created:             Wed Apr 03 11:46:43 +0800 (3 minutes ago)
Started:             Wed Apr 03 11:46:43 +0800 (3 minutes ago)
Finished:            Wed Apr 03 11:49:23 +0800 (49 seconds ago)
Duration:            2 minutes 40 seconds
Progress:            3/3
ResourcesDuration:   4m8s*(1 cpu),4m8s*(100Mi memory)
Parameters:          
  data_multiplier:   2000
  registry:          test-registry

STEP                                TEMPLATE                   PODNAME                                                     DURATION  MESSAGE
 ✔ machine-learning-pipeline-z72gm  machine-learning-pipeline                                                                          
 ├─✔ preprocess-data                preprocess-data            machine-learning-pipeline-z72gm-preprocess-data-4229626381  1m          
 ├─✔ train-data                     train-data                 machine-learning-pipeline-z72gm-train-data-1389575193       45s         
 └─✔ test-data                      test-data                  machine-learning-pipeline-z72gm-test-data-2535188255        13s
```

Before running the pipeline with vineyard, you need to enable the best effort scheduling for the vineyard runtime as follows:

```shell
# enable the fuse affinity scheduling
$ kubectl edit configmap webhook-plugins -n fluid-system
data:
  pluginsProfile: |
    pluginConfig:
    - args: |
        preferred:
          - name: fluid.io/fuse
            weight: 100
    ...

# Restart the fluid-webhook pod
$ kubectl delete pod -lcontrol-plane=fluid-webhook -n fluid-system
```

Then you can run the pipeline with vineyard via the following command:

```shell
$ argo submit pipeline-with-vineyard.yaml -p data_multiplier=2000 -p registry="test-registry"
Name:                machine-learning-pipeline-with-vineyard-q4tfr
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Pending
Created:             Wed Apr 03 12:00:45 +0800 (now)
Progress:            
Parameters:          
  data_multiplier:   2000
  registry:          test-registry
```

The result is shown as follows:

```shell
$ argo get machine-learning-pipeline-with-vineyard-q4tfr                               
Name:                machine-learning-pipeline-with-vineyard-q4tfr
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Succeeded
Conditions:          
 PodRunning          False
 Completed           True
Created:             Wed Apr 03 12:00:45 +0800 (2 minutes ago)
Started:             Wed Apr 03 12:00:45 +0800 (2 minutes ago)
Finished:            Wed Apr 03 12:02:36 +0800 (34 seconds ago)
Duration:            1 minute 51 seconds
Progress:            3/3
ResourcesDuration:   2m40s*(1 cpu),2m40s*(100Mi memory)
Parameters:          
  data_multiplier:   2000
  registry:          test-registry

STEP                                              TEMPLATE                                 PODNAME                                                                  DURATION  MESSAGE
 ✔ machine-learning-pipeline-with-vineyard-q4tfr  machine-learning-pipeline-with-vineyard                                                                                       
 ├─✔ preprocess-data                              preprocess-data                          machine-learning-pipeline-with-vineyard-q4tfr-preprocess-data-869469459  55s         
 ├─✔ train-data                                   train-data                               machine-learning-pipeline-with-vineyard-q4tfr-train-data-4177295571      26s         
 └─✔ test-data                                    test-data                                machine-learning-pipeline-with-vineyard-q4tfr-test-data-1755965473       8s
```

From the results, it's evident that the pipeline utilizing vineyard reduces the running time by approximately 30% compared to the pipeline without vineyard. This improvement is due to vineyard's ability to accelerate data sharing by leveraging shared memory, which is more efficient than traditional methods such as NFS.

### Modifications to use vineyard

Compared to the original kubeflow pipeline, we could use the following command to check the differences:

```shell
$ git diff --no-index --unified=40 pipeline.py pipeline-with-vineyard.py
```

The main modifications are:
- Add the vineyard persistent volume to the pipeline. This persistent volume is used to mount the vineyard client config file to the pipeline.

Also, you can check the modifications of the source code as 
follows.

- [Save data to vineyard](./preprocess-data/preprocess-data.py#L32-L35).
- [Load data from vineyard](./train-data/train-data.py#L15-L16).
- [load data from vineyard](./test-data/test-data.py#L14-L15).

The main modification is to use vineyard to load and save data
rather than using local files.
