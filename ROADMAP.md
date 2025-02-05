# Fluid Roadmap

## Fluid 2025 Roadmap

### **1. Data Anyway**  
**Objective**: Enable fluid data access **regardless of infrastructure constraints** (e.g., storage types, runtime environments) with minimal effort.

#### **Technical Directions**  
- **Unified Cache Runtime Framework**  
  - Enable integration of new cache runtimes(e.g., Cubefs, DragonFly) with minimal code changes.  
  - Standardize APIs for cache engine compatibility (e.g., Alluxio, Vineyard, JuiceFS).  
- **ThinRuntime Productization**:  
  - Support dynamic volume mounting capabilities for multi-cloud/hybrid-cloud scenarios.  
  - Improve stability and performance for large-scale deployments.  
  - Minimum container permission (remove the privileged permission of FUSE Pod)

### **2. Data Anywhere**  
**Objective**: Achieve **cross-region, cross-cluster, and cross-platform** data mobility and accessibility.  

#### **Technical Directions**  
- **Multi-Cluster Dataset Unified Management**  
  - **Global Dataset**: Create datasets pointing to the same data source across clusters.  
  - **Queue Integration**: Orchestrate dependencies between data preparation and task scheduling.  
- **Efficient Data Prewarming & Migration**  
  - **Distributed Prewarming**: Maximize bandwidth utilization for fast data loading.  
  - **Throttling Control**: Limit bandwidth usage during prewarming to avoid saturation.  
  - **Rsync Optimization**: Improve cross-region sync efficiency (e.g., Horizon integration).  
- **Intelligent Scaling**:  
  - Recommend Pods for scaling (e.g., prioritize underutilized nodes).  
  - Ensure cache engine compatibility with dynamic throughput adjustments post-scaling.  
- **Disk-Aware Scheduling**:  
  - Schedule workloads based on disk capacity and utilization.  
  - Support standard scenarios (e.g., disk-level resource allocation).  
- **Observability-Driven Optimization**  
  - **Pattern Recognition**: Analyze data access patterns to auto-inject acceleration components (e.g., caching, prefetching).  
  - **Idle Dataset Detection**: Identify unused datasets via reference counting and access history.  

---

### **3. Data Anytime**  
**Core Goal**: Ensure **real-time, adaptive, and intelligent** data availability for workloads.  

#### **Technical Directions**  

---



