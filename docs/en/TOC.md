# Fluid Documentation

<!-- markdownlint-disable MD007 -->
<!-- markdownlint-disable MD032 -->

## TOC

+ Overview
  - [Introduction](userguide/overview.md)
+ Get Started
  - [Quick Start](userguide/get_started.md)
  - [Installation](userguide/install.md)
  - [Trubleshooting](userguide/troubleshooting.md)
+ Dataset
  + Creation
    - [Accelerate Data Accessing(via POSIX)](samples/accelerate_data_accessing.md)
    - [Accelerate Data Accessing(via HDFS interface)](samples/accelerate_data_accessing_by_hdfs.md)
    - [Cache Co-locality](samples/data_co_locality.md)
    - [Share data across namespace (CSI mode)](samples/dataset_across_namespace_with_csi.md)
    - [Share data across namespace (Sidecar mode)](samples/dataset_across_namespace_with_sidecar.md)
  + Operation
    - [Data Preloading](samples/data_warmup.md)
    - [Cache Runtime Manually Scaling](samples/dataset_scaling.md)
  + Security
    - [Encrypted options for Dataset](samples/use_encryptoptions.md)
    - [Using Fluid to access non-root user's data](samples/nonroot_access.md)
    - [Set Access Mode](samples/data_accessmodes.md)
+ Storage
    - [Accelerate HostPath with Fluid](samples/hostpath.md)
    - [Accelerate PVC with Fluid](samples/accelerate_pvc.md)
    - [Accelerate different Storage with Fluid](samples/accelerate_different_storage.md)
+ Workload
  - [Machine Learning](samples/machinelearning.md)
+ Advanced   
  - [Accelerate Data Access by MEM or SSD](samples/accelerate_data_by_mem_or_ssd.md)
  - [Alluxio Tieredstore Configuration](samples/tieredstore_config.md)
  - [Pod Scheduling Optimization](operation/pod_schedule_optimization.md)
  - [Pod Scheduling Base on Runtime Tiered Locality](operation/tiered_locality_schedule.md)
  - [Set FUSE clean policy](samples/fuse_clean_policy.md)
  + Serverless
    - [How to run in Knative environment](samples/knative.md)
    - [How to ensure the completion of serverless tasks](samples/application_controller.md)
  - [How to enable FUSE auto-recovery](samples/fuse_recover.md)
  - [Using Fluid on ARM64 platform](samples/arm64.md)
  - [Support Image Pull Secrets](samples/image_pull_secrets.md)
+ Operation Guide
  - [Runtime monitoring](operation/monitoring.md)
  - [Cache Runtime Auto Scaling](operation/dataset_auto_scaling.md)
+ Troubleshooting
  - [Collecting logs](userguide/troubleshooting.md)
+ Developer Guide
  - [How to develop](dev/how_to_develop.md)
  - [API_Doc](dev/api_doc.md)
  - [Develop with Kind on MacOS](dev/dev_with_kind.md)
  - [Performance Analyze with pprof](dev/pprof.md)
  + Client Usage
    - [How to create and delete fluid resources using Go client](dev/use_go_create_resource.md)
    - [How to use client other than Go client](dev/multiple-client-support.md)
    - [Access via REST API](samples/api_proxy.md)
