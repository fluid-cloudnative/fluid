# Fluid Release Notes

## v0.6.0

### Features

- Support dataset cache autoscaling and cronscaling
- Add dataset mounting point dynamically update feature
- Enhance dataset cache aware Pod scheduling 
- Enhance HA support for cache Runtime
- Support new cache Runtime：GooseFS

### Bugs

- Fix [if alluxioruntime is nonroot, databackup will fail](https://github.com/fluid-cloudnative/fluid/issues/745)
- Fix [Node labels exceeds maximum length limit for long namespace and name](https://github.com/fluid-cloudnative/fluid/issues/704)


## v0.5.0

### Features

- Support on-the-fly dataset cache scale out/in
- Add Metadata backup and restore operation
- Support Fuse global mode and toleration
- Enhance Prometheus monitoring support for AlluxioRuntime
- Support new Runtime：JindoFS
- Support HDFS configuration

### Bugs

- [Fix compatibality issue of K8s 1.19+](https://github.com/fluid-cloudnative/fluid/issues/603)

## v0.4.0

### Features

- Warm up Dataset automatically before using it
- Support managing 4,000,000 files
- Support deploying multiple dataset in the same node
- Support showing the HCFS Access Endpoint in Dataset status

## v0.3.0

### Features

- Accelerate data access in Host-path mode in K8s
- Accelerate data access in Persistent Volume mode in K8s
- Support the underlay storage(NFS, Lustre) which is configured only with non-root
- Make the Alluxio Runtime’s settings optimized by default


### Bugs

- [MountVolume.SetUp failed in GKE](https://github.com/fluid-cloudnative/fluid/issues/222)
- [fuse.csi.fluid.io not found in the list registered CSI drivers when node restart in k8s 1.15.1](https://github.com/fluid-cloudnative/fluid/issues/220)

## v0.2.0

### Features

- Add code coverage badge  
- Update chart v0.2.0  
- Update docs  
- Refactor the package name  

## v0.1.0

### Features

- Add 'Dataset', 'AlluxioRuntime' CRD for managing the data
- Add PV and PVC for Posix interface
- Documentation
- Helm Chart
