# Introduction

Fluid provides Kubernetes users with a simple and efficient cloud data access scheme. At present, Fluid is compatible with multiple distributed cache engines through the Runtime Plugin. At present, it has supported JindoFS of Alibaba Cloud EMR, open-source Alluxio, Juicefs and other cache engines to facilitate users to interact with the underlying storage system through the above cache system and open the data access link.

In addition to the integrated cache engines above, users may need to access their own storage systems. In order to meet this requirement, users must complete the docking between the self built storage system and the Kubernetes environment by themselves. They need to write a customized CSI plug-in or a Runtime Controller based on the self built storage system to interface with Fluid, which requires users to have relevant knowledge of Kubernetes, thus increasing the work cost of the storage provider.



In order to meet users' needs to access data from other general storage systems using Fluid, the Fluid community has developed ThinRuntime to facilitate rapid access to other storage systems.

## Demo

| Storage System |                User Guide                | 
|----------------|:----------------------------------------:|
| NFS            |      [User Guide](./nfs/readme.md)       |
| CubeFS 2.4     |  [User Guide](./cubefs/v2.4/readme.md)   |
| CephFS         |     [User Guide](./cephfs/readme.md)     |
| GlusterFS      | [User Guide](./glusterfs/readme.md)      |
| Curvine        |      [User Guide](./curvine/readme.md)   |