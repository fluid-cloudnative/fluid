# Fluid Documentation

<!-- markdownlint-disable MD007 -->
<!-- markdownlint-disable MD032 -->

## TOC

+ 概述
  - [介绍](userguide/overview.md)
+ 入门
  - [安装](userguide/install.md)
  - [快速开始](userguide/get_started.md)
  - [问题诊断](userguide/troubleshooting.md)
+ 数据集使用
  + 创建
    - [数据加速（通过POSIX接口访问）](samples/accelerate_data_accessing.md)
    - [数据加速（通过HDFS接口访问）](samples/accelerate_data_accessing_by_hdfs.md)
    - [数据亲和性调度](samples/data_co_locality.md)
    - [数据容忍污点调度](samples/data_toleration.md)
    - [Fuse客户端全局部署](samples/fuse_affinity.md)
  + 操作
    - [数据预加载](samples/data_warmup.md)
    - [手动扩缩容](samples/dataset_scaling.md)
  + 安全
    - [使用参数加密](samples/use_encryptoptions.md)
    - [以non-root用户身份使用](samples/nonroot_access.md)
+ 底层存储
  - [主机目录加速](samples/hostpath.md)
  - [数据卷加速](samples/accelerate_pvc.md)
+ 进阶使用
  - [AlluxioRuntime分层存储配置](samples/tieredstore_config.md)
+ 工作负载
  - [机器学习](samples/machinelearning.md)
+ 更多Runtime实现
  - [使用 JindoRuntime](https://github.com/aliyun/alibabacloud-jindofs/blob/master/docs/jindo_fluid/jindo_fluid_overview.md)
+ 运维指南
  - [运行时监控](operation/monitoring.md)
  - [JVM性能分析](dev/profiling.md)
  - [自动弹性伸缩](operation/dataset_auto_scaling.md)
  - [定时弹性伸缩](operation/dataset_cron_scaling.md)
+ 开发者指南
  - [如何参与开发](dev/how_to_develop.md)
  - [API_Doc](dev/api_doc.md)
  - [如何增加新Runtime实现](dev/runtime_dev_guide.md)
  + 客户端使用
    - [如何使用Go客户端创建、删除fluid资源](dev/use_go_create_resource.md)
    - [如何使用其他客户端（非GO语言）](dev/multiple-client-support.md)
    - [通过REST API访问](samples/api_proxy.md)

