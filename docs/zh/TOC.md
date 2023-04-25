# Fluid Documentation

<!-- markdownlint-disable MD007 -->
<!-- markdownlint-disable MD032 -->

## TOC

+ 概述
  - [概述](userguide/overview.md)
+ 核心概念
  - [Fluid简介](core-concepts/introduction.md)
  - [概念](core-concepts/concepts.md)
  - [系统架构](core-concepts/architecture.md)
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
    - [跨namespace共享数据(CSI模式)](samples/dataset_across_namespace_with_csi.md)
    - [跨namespace共享数据(sidecar模式)](samples/dataset_across_namespace_with_sidecar.md)
  + 操作
    - [数据预加载](samples/data_warmup.md)
    - [Cache Runtime手动扩缩容](samples/dataset_scaling.md)
  + 安全
    - [使用参数加密](samples/use_encryptoptions.md)
    - [以non-root用户身份使用](samples/nonroot_access.md)
    - [修改访问模式](samples/data_accessmodes.md)
  + 高级配置
    - [配置元信息同步策略](samples/metadata_sync.md)
+ 底层存储
  - [主机目录加速](samples/hostpath.md)
  - [数据卷加速](samples/accelerate_pvc.md)
  - [加速不同的底层存储](samples/accelerate_different_storage.md)
+ 进阶使用
  - [使用内存加速和SSD加速配置](samples/accelerate_data_by_mem_or_ssd.md)
  - [AlluxioRuntime分层存储配置](samples/tieredstore_config.md)
  - [通过Webhook机制优化Pod调度](operation/pod_schedule_optimization.md)
  - [如何开启 FUSE 自动恢复能力](samples/fuse_recover.md)
  - [面向 ARM 架构的使用](samples/arm64.md)
  - [设置 FUSE 清理策略](samples/fuse_clean_policy.md)
  - [镜像拉取密钥](samples/image_pull_secrets.md)
  + 无服务器场景
    - [如何在Knative环境运行](samples/knative.md)
    - [如何保障 Serverless 任务顺利完成](samples/application_controller.md)
+ 工作负载
  - [机器学习](samples/machinelearning.md)
+ 更多Runtime实现
  - [使用 JindoRuntime](https://github.com/aliyun/alibabacloud-jindodata/blob/master/docs/user/3.x/jindo_fluid/jindo_fluid_overview.md)
  - [使用 JuiceFSRuntime](https://github.com/juicedata/juicefs)
    - [JuiceFS 开源环境搭建步骤](samples/juicefs/juicefs_setup.md)
    - [如何在 Fluid 中使用 JuiceFS](samples/juicefs/juicefs_runtime.md)
    - [JuiceFSRuntime Worker 的配置](samples/juicefs/juicefs_worker.md)
    - [JuiceFSRuntime Dataset 的配置](samples/juicefs/juicefs_dataset.md)
    - [JuiceFSRuntime 缓存配置](samples/juicefs/juicefs_cache_dir.md)
    - [JuiceFSRuntime 加速数据访问](samples/juicefs/juicefs_data_accelerate.md)
    - [JuiceFSRuntime 数据迁移](samples/juicefs/juicefs_data_migrate.md)
+ 运维指南
  - [运行时监控](operation/monitoring.md)
  - [JVM性能分析](dev/profiling.md)
  - [自动弹性伸缩](operation/dataset_auto_scaling.md)
  - [定时弹性伸缩](operation/dataset_cron_scaling.md)
  - [pprof性能分析](dev/pprof.md)
+ 问题诊断
  - [日志收集](userguide/troubleshooting.md)
  - [数据卷挂载问题](troubleshooting/debug-fuse.md)
+ 开发者指南
  - [如何参与开发](dev/how_to_develop.md)
  - [API_Doc](dev/api_doc.md)
  - [如何增加新Runtime实现](dev/runtime_dev_guide.md)
  + 客户端使用
    - [如何使用Go客户端创建、删除fluid资源](dev/use_go_create_resource.md)
    - [如何使用其他客户端（非GO语言）](dev/multiple-client-support.md)
    - [通过REST API访问](samples/api_proxy.md)

