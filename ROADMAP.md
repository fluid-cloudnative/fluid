# Fluid Roadmap
## 2022

### Make Fluid CSI Plugin ready for production:

- Introduce on-demand start-up mechanism for FUSE Pod: This reduces resource usage of cache engines in large-scale scenarios.
- Implement automatic recovery feature for Fuse mount points: This reduces the impact of Fuse process crashes on applications, improving system stability.

### Enable Fluid to Run Anywhere & Everywhere::

- Support Fluid FUSE client with both sidecar and CSI plugin mode with the configuration: Provides the flexibility for the deployment of Fluid.
- Support Job Terminator for the sidecar mode: This enhances the control over job execution.
- Support arm64 architecture: This broadens the platform compatibility of Fluid.
- Make Fluid's components compatible for different environments: Suppports not only native environments but also edge computing, Serverless Kubernetes, and multi-cluster Kubernetes setups.

### Expand Support for More Cache Runtimes:

- Add JuiceFS/JindoFS support: Expands the versatility of the Fluid framework to handle more types of runtimes and storage types.



