# Using Fluid in Kubeflow Pipelines v2
This is a demo that wraps Fluid'operations as KFP v2 components to complete the processes of dataset creation, runtime creation, and cache preheating to accelerate model training (a simple CNN model for [Fashion MNIST](https://www.kaggle.com/datasets/zalando-research/fashionmnist)).

## Prerequisites

### Installation 
- [Fluid](https://github.com/fluid-cloudnative/fluid)

- [Kubeflow Pipelines v2](https://www.kubeflow.org/docs/components/pipelines/v2/)

Please refer to the [Fluid installation guide](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) and [KFP v2 installation guide](https://www.kubeflow.org/docs/components/pipelines/v2/installation/quickstart/) to complete the installation of Fluid and KFP v2.

### Dataset
- [Fashion MNIST](https://www.kaggle.com/datasets/zalando-research/fashionmnist)

You should upload `fashion-mnist_train.csv` and `fashion-mnist_test.csv` to your Amazon S3 (or any S3-compatible storage, such as [MinIO](https://min.io/)) and deploy the [s3-secret.yaml](./s3-secret.yaml) to provide the access key.

### RBAC
Because KFP components require access or modification permissions to Fluid resources, it is necessary to deploy [rbac.yaml](./rbac.yaml) in advance to grant permissions.


## Demo
The sample is composed of dataset creation, alluxioruntime creation, data  preloading ,and model training.

The simple pipeline provides the following parameters:
- batch_size: int
- dataset_name: str
- epochs: int
- learning_rate: float
- mount_point: str
- mount_s3_endpoint: str
- mount_s3_region: str
- namespace: str (For now, this value should be the namespace in which you deploy your KFP)

If you want to run the sample, you can upload the [train-cnn-for-fashion-mnist-pipline.yaml](./pipline-yaml/train-cnn-for-fashion-mnist-pipline.yaml) to your pipeline dashboard UI and fill in these parameters.

Most importantly, this is just an example of packaging Fluid operations into KFP components, and users need to develop KFP components according to their own needs.