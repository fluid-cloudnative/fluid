# Fluid Alluxio Runtime 任务文档

## 目录
+ 数据访问
  - [数据加速（通过POSIX接口访问）](./dataset_usage/accelerate_data_accessing.md)
  - [数据加速（通过HDFS接口访问）](./dataset_usage/accelerate_data_accessing_by_hdfs.md)
  - [REST API访问数据](./dataset_usage/api_proxy.md)
+ 调度
  - [单机多DataSet使用配置](./dataset_schedule/multi_dataset_same_node_accessing.md)
  - [数据亲和性调度](./dataset_schedule/data_co_locality.md)
  - [数据容忍污点调度](./dataset_schedule/data_toleration.md)  
  - [通过Webhook机制优化Pod调度](./dataset_schedule/pod_schedule_global.md)
+ 预加载
  - [数据预加载](./dataload//data_warmup.md)
+ 备份和恢复
  - [数据备份与恢复](./databackup/data_backup_and_restore_metadata.md)
+ 弹性伸缩
  - [手动扩缩容](./dataset_scaling/dataset_scaling.md)
  - [自动弹性伸缩](./dataset_scaling/dataset_auto_scaling.md)
  - [定时弹性伸缩](./dataset_scaling/dataset_cron_scaling.md)
+ 底层存储系统
  - [主机目录加速](./under_storage/hostpath.md)
  - [数据卷加速](./under_storage/accelerate_pvc.md)
  - [HDFS存储系统](./under_storage/hdfs_configuration.md)
  - [S3存储系统](./under_storage/s3_configuration.md)
  - [GCS存储系统](./under_storage/gcs_configuration.md)
+ 可靠性
  - [Runtime Master高可用](./high_avaliable/master_high_availiable.md)
  - [如何开启 FUSE 自动恢复能力](./high_avaliable/fuse_recover.md)
+ 安全
  - [使用参数加密](./security/use_encryptoptions.md)
  - [以non-root用户身份使用](./security/nonroot_access.md)
+ 无服务器场景
  - [如何在Knative环境运行](./serverless/knative.md)
  - [如何保障 Serverless 任务顺利完成](./serverless/application_controller.md)
+ Runtime 进阶配置
  - [分层存储](./configuration/tieredstore_config.md)
+ 场景示例
  - [机器学习](./examples/machinelearning.md)
