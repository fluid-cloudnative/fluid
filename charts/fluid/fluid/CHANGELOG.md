### 0.1.0

* Update CSI image
* Update dataset CRD

### 0.2.0

* Refactor and clean up the code


### 0.3.0

* Speed up Volume and hostPath in Kubernetes
* Add RunAs


### 0.4.0

* Add debug info for csi
* Make mount root configurable
* Update HCFS URL
* Split the controller into alluxio runtime and dataset
* Implement DataLoad CRD and DataLoad controller


### 0.5.0

* Remove hostnetwork from the controller config
* Add JindoRuntime
* Avoid running in virtual kubelet node

### 0.6.0

* Add data affinity scheduling
* Auto Scaling
* High Availability
* Update mountPoint dynamically in runtime
* Add GooseFSRuntime

### 0.7.0

* Add mountPropagation for registrar
* Add syncRetryDuration
* Add auto fuse recovery

### 0.8.0

* Add application controller component
* Add Go gull profile capablities
* Support setting global image pull secrets
* Update mutating webhook configuration rules
* Support configurable pod metadata of runtimes
* Scale runtime controllers on demand

### 0.9.0
* Support pass image pull secrets from fluid charts to alluxioruntime controller
* Fix components rbacs and set Fluid CSI Plugin with node-authorized kube-client

### 0.9.1
* Fix CSI Plugin loop mount bug
