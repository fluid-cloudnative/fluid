# Fluid Documentation

<!-- markdownlint-disable MD007 -->
<!-- markdownlint-disable MD032 -->

## TOC

+ Overview
  - [Introduction](userguide/overview.md)
+ Get Started
  - [Quick Start](userguide/get_started.md)
  - [Installation](userguide/install.md)
  - [Troubleshooting](userguide/troubleshooting.md)
+ How To Use
  + Creation
    - [Accelerate Data Accessing(via POSIX)](samples/accelerate_data_accessing.md)
    - [Accelerate Data Accessing(via HDFS interface)](samples/accelerate_data_accessing_by_hdfs.md)
    - [Cache Co-locality](samples/data_co_locality.md)
  + Operation
    - [Data Preloading](samples/data_warmup.md)
  + Security
    - [Encrypted options for Dataset](samples/use_encryptoptions.md)
    - [Using Fluid to access non-root user's data](samples/nonroot_access.md)
+ Storage
    - [Accelerate HostPath with Fluid](samples/hostpath.md)
    - [Accelerate PVC with Fluid](samples/accelerate_pvc.md)
+ Workload
  - [Machine Learning](samples/machinelearning.md)
+ Advanced   
  - [Alluxio Tieredstore Configuration](samples/tieredstore_config.md)
+ Developer Guide
  - [How to develop](dev/how_to_develop.md)
  - [API_Doc](dev/api_doc.md)
  - [Develop with Kind on MacOS](dev/dev_with_kind.md)
  - [How to use client other than Go client](dev/multiple-client-support.md)
