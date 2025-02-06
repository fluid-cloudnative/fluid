# Fluid Roadmap

## Fluid 2025 Roadmap

### **1. Data Anyway**  
**Objective**: Enable fluid data access **regardless of infrastructure constraints** (e.g., storage types, runtime environments) without developing controller code.

- **Unified Cache Runtime Framework**  
  - Enable integration of new cache runtimes(e.g., Cubefs, DragonFly) via a **generic Cache Runtime interface** with minimal code changes.    
  - Standardize APIs for cache engine compatibility (e.g., Alluxio, Vineyard, JuiceFS).  
- **Adaptive Data Access**:  
  - Data Access Mode based on Scheduler's Decsion:  
    - *Shared-Kernel Nodes* → Use CSI plugins for direct mounting.  
    - *Kata Containers* → Switch to sidecar-based container. 
- **ThinRuntime Productization**:  
  - Improve stability and performance for large-scale deployments.  
  - Minimum container permission (remove the privileged permission of FUSE Pod)


### **2. Data Anywhere**  
**Objective**: Achieve **cross-region, cross-cluster, and cross-platform** data mobility and accessibility.  

- **Multi-Cluster Dataset Unified Management**  
  - **Global Dataset**: Create datasets pointing to the same data source across clusters.  
  - **Queue Integration**: Orchestrate dependencies between data preparation and task scheduling.  
  - **Persistent Data Mirroring**  
    - **Region-Aware Replication**: Automatically mirror datasets across clouds/regions.  
    - **Consistency Guarantees**: Support both eventual and strong consistency models.  

- **Efficient Data Prewarming & Migration**  
  - **Distributed Prewarming**: Maximize bandwidth utilization for fast data loading.  
  - **Throttling Control**: Limit bandwidth usage during prewarming to avoid saturation.  
  - **Rsync Optimization**: Improve cross-region sync efficiency.  

- **Elastic Caching & Scheduling**:  
  - **Disk-Aware Scheduling**: Optimize workload placement based on disk capacity, utilization, and locality.  
  - **Intelligent Scaling**:  
    - Recommend underutilized Pods for scaling (cost/performance-aware).  
    - Ensure cache engines adapt to dynamic throughput post-scaling.  
  - **Cloud-Agnostic Recovery**: Rebuild caches across regions using cloud disk snapshots.  

- **Observability-Driven Optimization**  
  - **Pattern Recognition**: Analyze data access patterns to auto-inject acceleration components (e.g., caching, prefetching).  
  - **Idle Dataset Detection**: Identify unused datasets via reference counting and access history.  

- **Application-Side Acceleration**  
  - **Transparent Prefetching**:  
    - Inject sidecar containers to prefetch data dynamically (e.g., Alluxio/Fluid Runtime).  
    - Auto-adjust prefetch strategies (block size, concurrency) based on access patterns.  
  - **Dynamic SDK Injection**: Attach acceleration SDKs to Pods via Fluid Admission Controller (no base image modification).  


### **3. Data Anytime**  
**Core Goal**: Ensure **real-time, adaptive, and intelligent** data availability for workloads.  

- **Temporal Workflows with Kueue**:  
  - Trigger ML jobs (TFJob, PyTorchJob) **after prewarming completes**.  
  - Automate post-job cleanup (data migration/cache eviction).  
- **Dynamic Volume Mounting**:  
  - Support dynamic volume mounting capabilities for multi-cloud/hybrid-cloud scenarios.  
  - Enable dyanmic data mount operations in Python SDK. 

