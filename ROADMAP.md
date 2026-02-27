# Fluid Roadmap

## Fluid 2026 Roadmap

### **1. Data Anyway**  
**Objective**: Enable fluid data access **regardless of infrastructure constraints** (e.g., storage types, runtime environments) without developing controller code.

- **Generic Cache Runtime**    
  – **Pluggable Architecture**: Standardized Cache Runtime Interface for rapid integration of new engines (CubeFS, Dragonfly, Vineyard) with minimal boilerplate.
  – **AdvancedStatefulset-Based Orchestration**: Migrate from StatefulSet to InstanceSet for fine-grained Pod lifecycle management, ordered rollout, and enhanced failover capabilities.

- **Runtime Dynamic Configuration**

  – **Zero-Downtime Tuning**: Adjust cache replicas, storage media tiers (SSD/HDD/DRAM), and eviction policies without Dataset reconstruction or workload restart.

  – **Hot Parameter Swapping**: Runtime modification of cache engine configurations (e.g., Alluxio block size, Jindo worker threads) for traffic spike handling.

- **API upgradation to v1alpha2**

  – Standardized Conditions, ObservedGeneration, and phase transition semantics for improved GitOps and tooling compatibility.

  – Conversion webhook support for seamless v1alpha1 → v1alpha2 migration.

- **Validation Webhook**

 – Admission-time CRD validation with auto-correction suggestions to prevent misconfigurations.

 – Policy enforcement for resource quotas and security constraints.

- **ThinRuntime Productization**

 – Production-ready stability for large-scale deployments with **minimum container privileges** (eliminate privileged FUSE Pod requirements).

### **2. Data Anywhere**  
**Objective**: Achieve **cross-region, cross-cluster, and cross-platform** data mobility and accessibility.  

- **LLM KV Cache Orchestration**_(P0, New Strategic Focus)_

 – **Disaggregated KV Cache**: Externalize vLLM/SGLang KV Cache to Fluid-managed distributed storage, enabling 10x+ throughput improvement for long-context inference.

 – **Cross-Pod Cache Sharing**: Live migration of KV Cache between inference instances for preemptive scheduling and spot instance tolerance.

 – **Mooncake Integration**: Official partnership for high-performance KV Cache backend with RDMA acceleration.


- **Efficient Data Prewarming & Migration**  
  - **Distributed Prewarming**: Maximize bandwidth utilization for fast data loading.  
  - **Throttling Control**: Limit bandwidth usage during prewarming to avoid saturation.  
  - **Rsync Optimization**: Improve cross-region sync efficiency.  

 - **Elastic Caching & Cost Optimization**

  – **Disk-Aware Scheduling**: Integrate node disk capacity, utilization, and data locality into Kubernetes scheduler decisions.

  – **Kubernetes Node-Specific Restrictions**_(P1)_: Support node selectors, affinity/anti-affinity, and taint tolerations for cache Pods.

 – **Intelligent Autoscaling**: Cost-performance-aware recommendations with cache engine throughput adaptation post-scaling.

 – **Cloud-Native Recovery**: Cross-region cache reconstruction via cloud disk snapshots (EBS/Cloud Disk).

- **JindoRuntime High Availability**

 – **Master Pod Crash Recovery**: Automatic re-setup and state reconstruction after cache master failure without data loss.

 – **Metadata Persistence**: WAL-based metadata recovery for rapid failover.

- **Observability-Driven Optimization**

 – **Access Pattern Recognition**: ML-based analysis to auto-inject acceleration strategies (prefetching, block size optimization).

 – **Dataset Garbage Collection**: Idle dataset detection via reference counting and access history analysis. 


### **3. Data Anytime**  
**Core Goal**: Ensure **real-time, adaptive, and intelligent** data availability for workloads.  

• **Temporal Workflow Integration**

 – **Kueue-Driven Pipelines**: Trigger training/inference jobs automatically upon DataLoad completion; automate post-job cache eviction and data migration.

 – **Event-Driven Policies**: Flexible metadata synchronization triggered by workload lifecycle events.

- **Developer Experience**

 – **Fluid kubectl Plugin**_(P1)_: Native CLI extension (kubectl fluid) for:

  -  Dataset status inspection and health diagnostics. 
  - On-demand prewarming triggering (kubectl fluid warmup). 
  - Cache performance profiling and bottleneck analysis. 
  - Runtime configuration hot-updates. 
