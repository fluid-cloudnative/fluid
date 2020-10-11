# Fluid Release Notes

## v0.1.0

### Features

- Add 'Dataset', 'AlluxioRuntime' CRD for managing the data
- Add PV and PVC for Posix interface
- Documentation
- Helm Chart


## v0.2.0

### Features

- Add code coverage badge  
- Update chart v0.2.0  
- Update docs  
- Refactor the package name  


## v0.3.0

### Features

- Accelerate data access in Host-path mode in K8s
- Accelerate data access in Persistent Volume mode in K8s
- Support the underlay storage(NFS, Lustre) which is configured only with non-root
- Make the Alluxio Runtime’s settings optimized by default



## v0.4.0

### Features


### Bugs

- [MountVolume.SetUp failed in GKE](https://github.com/fluid-cloudnative/fluid/issues/222)
- [fuse.csi.fluid.io not found in the list registered CSI drivers when node restart in k8s 1.15.1](https://github.com/fluid-cloudnative/fluid/issues/220)