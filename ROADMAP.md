# Fluid Roadmap
## 2021
### Optimize Data Access Acceleration
Objective: "Speedup data access for diversified scenarios on cloud"
* Support dataset metadata backup and restore （for small files）
* Support more cache runtimes (eg. Vineyard) for diversified data types
* Optimize cache runtime for the running apps on-the-fly

### Refine Data/App Scheduling
Objective: "Make cache runtime and application scheduling intelligent"
* Flexible scale in/out capability of cache runtimes (HPA)
* Workload-specific data cache orchestration
* Intelligent Data Prefetch based on Workload Scheduling History 

### Improve User Experience
Objective: "Enable operation and maintenance of Fluid with less cost"
* Data Abstraction: More data abstraction types for different computing frameworks
* Observability: Enhance observability for cache runtimes
* More Operations: Enhance Dataset Operation (Pin/Load/Free..)
* Data Management: Dataset Cache Garbage Collection in Kubernetes

