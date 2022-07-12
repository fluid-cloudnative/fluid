# Fluid Documentation

<!-- markdownlint-disable MD007 -->
<!-- markdownlint-disable MD032 -->

## TOC

+ 概念
  - [概述](concepts/overview.md)
  - [架构](concepts/architecture.md)
+ 安装
  - [入门](installation/get_started.md)
  - [安装](installation/installation.md)
  - [升级](installation/upgrade.md)
+ 任务
  + 选择不同运行时
  	- [Alluxio](./runtime_alluxio/task/README.md)
  	- [GooseFS]()
  	- [Jindo]()
  	- [JuiceFS]()
+ 运维指南
  - [安装和使用问题解答](./operation/faq.md)
  + 选择不同运行时
  	- [Alluxio](./runtime_alluxio/operation/README.md)
  	- [GooseFS]()
  	- [Jindo]()
  	- [JuiceFS]()
  - [日志收集](operation/collect_log.md)
  - [数据卷挂载问题](operation/debug-fuse.md)
+ 开发者指南
  - [如何参与开发](dev/how_to_develop.md)
  - [API_Doc](dev/api_doc.md)
  - [如何增加新Runtime实现](dev/runtime_dev_guide.md)
  + 客户端使用
    - [如何使用Go客户端创建、删除fluid资源](dev/use_go_create_resource.md)
    - [如何使用其他客户端（非GO语言）](dev/multiple-client-support.md)