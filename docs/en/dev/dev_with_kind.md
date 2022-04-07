# Fluid Development Environment with Kind on MacOS 

# Requirements
+ kind (version > v0.10.0)
+ docker 
+ MacOS

`csi-driver` and `node-driver-registrar`(sidecar) need to communicate with kubelet,   
so run them in kind container

# Set up steps
## 1. download go source code
In order to run `csi-driver` and `node-driver-registrar` code in kind container    
download it from [go1.16.3.linux-amd64.tar.gz](https://golang.org/dl/go1.16.3.linux-amd64.tar.gz)
decompress it and put it in `~/go/local/` directory

```
  ~/go
    |__bin
    |__pkg
    |__src
    |   |__ github.com
    |   |__ sigs.k8s.io
    |   |__ k8s.io
    |__local
    |  |__go 
```

## 2. create kind cluster 
```shell
$ kind create cluster --config=cluster.yaml
```
```
# cluster.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: fluid-dev
nodes:
- role: control-plane
  image: kindest/node:v1.16.1
  extraMounts:
  - hostPath: /your/Mac/go/path # GOPATH value on Mac
    containerPath: /home/work/go
```

```shell
# check kind cluster node
$ kubectl get nodes
```
```
NAME                      STATUS   ROLES    AGE   VERSION
fluid-dev-control-plane   Ready    master   74s   v1.16.15
```
## 3. create fluid crds and csi driver to kind cluster
+ alluxioruntimes.data.fluid.io
+ databackups.data.fluid.io
+ dataloads.data.fluid.io
+ datasets.data.fluid.io
+ jindoruntimes.data.fluid.io
+ [fuse.csi.fluid.io](https://github.com/fluid-cloudnative/fluid/blob/master/charts/fluid/fluid/templates/csi/driver.yaml)

```shell
$ cd ~/go/src/github.com/fluid-cloudnative/fluid
$ kustomize build config/crd/ | kubectl apply -f -
```
```
customresourcedefinition.apiextensions.k8s.io/alluxioruntimes.data.fluid.io created
customresourcedefinition.apiextensions.k8s.io/databackups.data.fluid.io created
customresourcedefinition.apiextensions.k8s.io/dataloads.data.fluid.io created
customresourcedefinition.apiextensions.k8s.io/datasets.data.fluid.io created
customresourcedefinition.apiextensions.k8s.io/jindoruntimes.data.fluid.io created
```
```shell
$ kubectl apply -f ./csi/deploy/csi-fluid-driver.yaml
$ kubectl get csidriver
```
```
NAME                CREATED AT
fuse.csi.fluid.io   2021-04-24T15:30:38Z
```

## 4.  download helm and create soft link charts
```shell
# download helm to ~/tmp/ unpack it and move to ~/go/bin
$ cd ~/tmp/ && curl https://get.helm.sh/helm-v3.6.2-darwin-amd64.tar.gz -o helm-v3.6.2-darwin-amd64.tar.gz
$ tar xzvf helm-v3.6.2-darwin-amd64.tar.gz  && cp darwin-amd64/helm ~/go/bin/ddc-helm

# create charts soft link
$ ln -s ~/go/src/github.com/fluid-cloudnative/fluid/charts ~/charts 
```
## 5. run csi-driver and node-driver-registrar code in kind node container

start fluid csi driver

```
docker exec -ti fluid-dev-control-plane /bin/bash
cd /home/work/go/src/github.com/fluid-cloudnative/fluid/cmd/csi && sh start.sh
```

```
#! /bin/bash
set -x

export TMPDIR=/root/go/tmp
export GO111MODULE=on
export GOMODCACHE=/root/go/pkg/mod
export GOPROXY=https://goproxy.io
export GOPATH=/home/work/go
export GOROOT=/home/work/go/local/go
export GOBIN=/home/work/go/bin
export PATH=$PATH:$GOBIN:$GOROOT/bin

if [ ! -d $TMPDIR ]; then
  mkdir -p $TMPDIR
fi


rm -f /var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock
mkdir -p /var/lib/kubelet/csi-plugins/fuse.csi.fluid.io

cp /home/work/go/src/github.com/allenhaozi/fluid/csi/shell/check_mount.sh /usr/local/bin/check_mount.sh && chmod +x /usr/local/bin/check_mount.sh 

/home/work/go/local/go/bin/go run main.go start \
	--endpoint="unix://var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock" \
    --nodeid="fluid-dev-control-plane" \
	--kubeconfig=/etc/kubernetes/kubelet.conf \
	--v=5
```
start node-driver-registrar
> I made a small change to [node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar/tree/v1.3.0) based on v1.3.0,    
  It can be downloaded from [fluid-dev-v1.3.0](https://github.com/allenhaozi/node-driver-registrar/tree/feat/fluid-dev-v1.3.0)  
  The only change is to allow passing `reg-path` in arguments   
  Does not affect the csi-driver online
```
docker exec -ti fluid-dev-control-plane /bin/bash
$ cd /home/work/go/src/github.com/allenhaozi/node-driver-registrar/cmd/csi-node-driver-registrar && sh start.sh

```

```
#! /bin/bash
set -x

export TMPDIR=/root/go/tmp
export GO111MODULE=on
export GOMODCACHE=/root/go/pkg/mod
export GOPROXY=https://goproxy.io
export GOPATH=/home/work/go
export GOROOT=/home/work/go/local/go
export GOBIN=/home/work/go/bin
export PATH=$PATH:$GOBIN:$GOROOT/bin

if [ ! -d $TMPDIR ]; then
  mkdir -p $TMPDIR
fi


# delete reg socket if exist
rm -rf /var/lib/kubelet/plugins_registry/fuse.csi.fluid.io-reg.sock

go run main.go \
	--kubelet-registration-path="/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock" \
	--csi-address="/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock" \
    --reg-path="/var/lib/kubelet/plugins_registry" \
    --v=5
```

## 6. start alluxio and dataset on Mac

```shell
cd ~/go/src/github.com/fluid-cloudnative/fluid/cmd/alluxio && sh start.sh
```

```
###
### start.sh
###
#! /bin/bash

export ALLUXIO_INIT_IMAGE_ENV=registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.4.0-a8ba7c9
export MOUNT_ROOT=/alluxio-mnt
go run main.go start

```

```shell
cd ~/go/src/github.com/fluid-cloudnative/fluid/cmd/dataset && sh start.sh
```

```
###
### start.sh
###
#! /bin/bash

export ALLUXIO_INIT_IMAGE_ENV=registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.4.0-a8ba7c9
go run main.go start \
   --metrics-addr=127.0.0.1:8082
```

## 7. load image to kind cluster
```
kind load docker-image registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.4.0-a8ba7c9 --name=fluid-dev
kind load docker-image registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio:2.3.0-SNAPSHOT-2c41226 --name=fluid-dev
kind load docker-image registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse:2.3.0-SNAPSHOT-2c41226 --name=fluid-dev                                 
kind load docker-image nginx:latest --name=fluid-dev
```

## 8. run fluid demo
```shell
$ kubectl apply -f dataset.yaml
$ kubectl apply -f runtime.yaml
$ kubectl apply -f pod.yaml
```

```
# check pod list
NAME                READY   STATUS    RESTARTS   AGE
demo-app            1/1     Running   0          31m
demo-fuse-ksdsr     1/1     Running   0          31m
demo-master-0       2/2     Running   0          36m
demo-worker-k2xhh   2/2     Running   0          31m
```
```shell
# check data
$ kubectl exec -ti demo-app sh
$ ls /data/spark/

SparkR_3.0.2.tar.gz   spark-3.0.2-bin-hadoop2.7-hive1.2.tgz  spark-3.0.2-bin-hadoop3.2.tgz  spark-3.0.2.tgz
pyspark-3.0.2.tar.gz  spark-3.0.2-bin-hadoop2.7.tgz      spark-3.0.2-bin-without-hadoop.tgz
```
