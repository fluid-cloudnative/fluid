# Developer Guide

## Requirements

- Git
- Golang (version >= 1.13)
- Docker (version >= 19.03)
- Kubernetes (version >= 1.14)
- GNU Make

For installation of Golang, please refer to [Install Golang](https://golang.org/dl/)

`make` is usually in a `build-essential` package in your distribution's package manager of choice. Make sure you have `make` on your machine.

There're great chances that you may want to run your implementation in a real Kubernetes cluster, so probably a Docker is needed for some necessary operations like building images.
See [Install Docker](https://docs.docker.com/engine/install/) for more information.

## How to Build, Run and Debug

### Get Source Code

```shell
$ mkdir -p $GOPATH/src/github.com/fluid-cloudnative/
$ cd $GOPATH/src/github.com/fluid-cloudnative
$ git clone https://github.com/fluid-cloudnative/fluid.git
```

> **NOTE**: In this document, we build, run and debug under non-module environment. 
>
> See [Go Modules](https://github.com/golang/go/wiki/Modules) for more information if some issue occurs to you.

### Build Binary
`Makefile` under project directory provides many tasks you may want to use including Test, Build, Debug, Deploy etc.

You can simply get a binary by running:
```shell
# build dataset-controller, alluxioruntime-controller and csi Binary
$ make build
```
By default, the binary would be put under `<fluid-path>/bin`.

### Build Images
1. Set tags for images
    
    ```shell
    # set name for image of dataset-controller
    $ export DATASET_CONTROLLER_IMG=<your-registry>/<your-namespace>/<img-name>
    # set name for image of alluxioruntime-controller
    $ export ALLUXIORUNTIME_CONTROLLER_IMG=<your-registry>/<your-namespace>/<img-name>
    # set name for image of  CSI
    $ export CSI_IMG=<your-registry>/<your-namespace>/<csi-img-name>
    # set name for image of init-user
    $ export INIT_USERS_IMG=<your-registry>/<your-namespace>/<csi-img-name>
    
    # build all images
    $ make docker-build-all
    ```
    Before running Fluid, you need to push the built image to an accessible image registry.

2. Login to a image registry
    
    Make sure you've login to a docker image registry that you'd like to push your image to:
    ```shell
    $ sudo docker login <docker-registry>
    ```

3. push your images:

    ```shell          
    $ make docker-push-all
    ```

### Run Your Fluid on Kubernetes Cluster
In the following steps, we assume you have properly configured `KUBECONFIG` environment variable or set up `~/.kube/config`. See [Kubeconfig docs](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) for more information.

1. Push your images to a image registry accessible to your Kubernetes cluster

    If your images are pushed to some private repositories, make sure your Kubernetes cluster hold credentials for accessing those repositories.

2. Change image  in the samples we provide:

    ```yaml
    # <fluid-path>/config/fluid/patches/image_in_manager.yaml
    ...
    ...
    containers:
      - name: manager
        image: <registry>/<namespace>/<img-repo>:<img-tag>
    ```
    ```yaml
    # <fluid-path>/config/fluid/patches/image_in_csi-plugin.yaml
    ...
    ...
    containers:
      - name: plugins
        image: <registry>/<namespace>/<csi-img-name>:<csi-img-tag>
    ```

3. Install CRDs
    ```shell
    $ kubectl apply -k config/crd
    ```
    
    Check CRD with:
    
    ```shell
    $ kubectl get crd | grep fluid
    alluxiodataloads.data.fluid.io          2020-08-22T03:53:46Z
    alluxioruntimes.data.fluid.io           2020-08-22T03:53:46Z
    datasets.data.fluid.io                  2020-08-22T03:53:46Z
    ```

4. Install your implementation
    ```shell
    $ kubectl apply -k config/fluid
    ```
    
    Check Fluid system with:
    
    ```shell
    $ kubectl get pod -n fluid-system
    NAME                                  READY   STATUS    RESTARTS   AGE
    controller-manager-7fd6457ccf-p7j2x   1/1     Running   0          84s
    csi-nodeplugin-fluid-pj9tv            2/2     Running   0          84s
    csi-nodeplugin-fluid-t8ctj            2/2     Running   0          84s
    ```

5. Run samples to verify your implementation

    Here is a sample provided by us, you may want to rewrite it according to your implementation.
    ```shell
    $ kubectl apply -k config/samples
    ```
    
    Check sample pods:
    
    ```shell
    $ kubectl get pod
    NAME                   READY   STATUS    RESTARTS   AGE
    cifar10-fuse-vb6l4     1/1     Running   0          6m15s
    cifar10-fuse-vtqpx     1/1     Running   0          6m15s
    cifar10-master-0       2/2     Running   0          8m24s
    cifar10-worker-729xz   2/2     Running   0          6m15s
    cifar10-worker-d6kmd   2/2     Running   0          6m15s
    nginx-0                1/1     Running   0          8m30s
    nginx-1                1/1     Running   0          8m30s
    ```

6. Check logs to verify your implementation
    ```shell
    $ kubectl logs -n fluid-system <controller_manager_name>
    ```

7. Clean up
    ```shell
    $ kubectl delete -k config/samples
    $ kubectl delete -k config/fluid
    $ kubectl delete -k config/crd
    ```

### Unit Testing

#### Basic Tests

Execute following command from project root to run basic unit tests:

```shell
$ make unit-test
```

#### Integration Tests

`kubebuilder` provided a integration test framework based on [envtest](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/envtest) package. You must install `kubebuilder` before running integration tests:

```shell
$ os=$(go env GOOS)
$ arch=$(go env GOARCH)
$ curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/
$ sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
$ export PATH=$PATH:/usr/local/kubebuilder/bin
```

Next, run all unit tests (integration tests included) with:

```shell
$ make test
```

> **NOTE:** When running unit tests on a non-linux system such as macOS, if testing failed and says `exec format error`, you may need to check whether `GOOS` is consistent with your actual OS.

### Debug

You can debug your program in multiple ways, here is just a brief guide for how to debug with `go-delve`

**Prerequisites**

Make sure you have `go-delve` installed. See [go-delve installation guide](https://github.com/go-delve/delve/tree/master/Documentation/installation) for more information

**Debug locally**
```shell
# build & debug in one line
$ dlv debug <fluid-path>/cmd/controller/main.go

# debug binary
$ make manager
$ dlv exec bin/manager
```

**Debug remotely**

On remote host:
```shell
$ dlv debug --headless --listen ":12345" --log --api-version=2 cmd/controller/main.go
```
The command above will make `go-delve` start a debug service and listen for port 12345.

On local host, connect to the debug service:
```shell
$ dlv connect "<remote-address>:12345" --api-version=2
```

> Note: To debug remotely, make sure the specified port is not occupied and the firewall has been properly configured.
