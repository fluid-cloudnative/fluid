# 介绍
Fluid 为 Kubernetes 用户提供一套简单高效的云上数据访问方案。目前，Fluid 通过 Runtime Plugin 的方式，扩展兼容多种分布式缓存引擎。目前已经支持了开源的 Alluxio，Juicefs 以及阿里云 EMR 的 JindoFS等缓存引擎，方便用户通过上述缓存系统与底层存储系统交互，打通数据访问链路。

除了上述已经集成的缓存引擎外，用户也有访问自建存储系统的需求。为满足这个需求，用户须自行完成自建存储系统与 Kubernetes 环境的对接工作，需要根据自建存储系统编写自定义的 CSI 插件或编写 Runtime Controller 与 Fluid 对接，这都要求用户具备 Kubernetes 的相关知识，从而增加了存储提供者的工作成本。

为了满足用户使用 Fluid 访问其他通用存储系统数据需求，Fluid 社区开发了 ThinRuntime 以方便其他存储系统的快速接入。

## 接入案例

| 存储系统         |                 文档                  | 
|--------------|:-----------------------------------:|
| NFS          |     [案例](./nfs/readme-zh_CN.md)     |
| CubeFS 2.4   | [案例](./cubefs/v2.4/readme-zh_CN.md) |
| CephFS       |   [案例](./cephfs/readme-zh_CN.md)    |
| GlusterFS    |  [案例](./glusterfs/readme-zh_CN.md)  |
| Curvine      |  [案例](./curvine/readme-zh_CN.md)    |
