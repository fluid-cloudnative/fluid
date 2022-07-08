# Fluid Alluxio Runtime 文档

## 目录

+ 概述
  - [介绍](../overview.md)
+ 入门
  - [安装](../install.md)
  - [快速开始](userguide/get_started.md)
+ 数据集
  + 使用
    - [数据加速（通过POSIX接口访问）](samples/accelerate_data_accessing.md)
    - [数据加速（通过HDFS接口访问）](samples/accelerate_data_accessing_by_hdfs.md)
    - [主机目录加速](samples/hostpath.md)
    - [数据卷加速](samples/accelerate_pvc.md)
    - [REST API访问数据](samples/api_proxy.md)
  + 调度
    - [数据亲和性调度](samples/data_co_locality.md)
    - [数据容忍污点调度](samples/data_toleration.md)  
    - [通过Webhook机制优化Pod调度](samples/pod_schedule_global.md)
  + 预加载
    - [数据预加载](samples/data_warmup.md)
  + 备份和恢复
    - [数据预加载](samples/backup_and_restore_metadata.md)
+ 进阶使用
  - [单机多DataSet使用配置](samples/multi_dataset_same_node_accessing.md)
  + 弹性伸缩
    - [手动扩缩容](samples/dataset_scaling.md)
    - [自动弹性伸缩](samples/dataset_auto_scaling.md)
    - [定时弹性伸缩](samples/dataset_cron_scaling.md)
  + 底层存储系统
    - [GCS存储系统](samples/gcs_configuration.md)
    - [HDFS存储系统](samples/hdfs_configuration.md)
    - [S3存储系统](samples/s3_configuration.md)
    - [OSS存储系统](samples/oss_configuration.md)
  + 可靠性
    - [Runtime Master高可用](samples/master_high_aviliability.md)
    - [如何开启 FUSE 自动恢复能力](samples/fuse_recover.md)
  + 安全
    - [使用参数加密](samples/use_encryptoptions.md)
    - [以non-root用户身份使用](samples/nonroot_access.md)
  + 无服务器场景
    - [如何在Knative环境运行](samples/knative.md)
    - [如何保障 Serverless 任务顺利完成](samples/application_controller.md)
  + Runtime 进阶配置
    - [分层存储](samples/tieredstore_config.md)
+ 场景示例
  - [机器学习](examples/machinelearning.md)
+ 运维指南
  - [运行时监控](operation/monitoring.md)
  - [JVM性能分析](dev/profiling.md)
+ 问题诊断
  - [FAQ](userguide/faq.md)
  - [日志收集](userguide/troubleshooting.md)
  - [数据卷挂载问题](userguide/debug-fuse.md)
+ 开发者指南
  - [如何参与开发](dev/how_to_develop.md)
  - [API_Doc](dev/api_doc.md)
  - [如何增加新Runtime实现](dev/runtime_dev_guide.md)
  + 客户端使用
    - [如何使用Go客户端创建、删除fluid资源](dev/use_go_create_resource.md)
    - [如何使用其他客户端（非GO语言）](dev/multiple-client-support.md)

