# Fluid Roadmap
## 2024

### Objective: Achieve orchestration of data operations and Kubernetes job scheduling systems

- Support temporality through Kueue
   - Once data migration is completed, run data preheating, triggering the running of machine learning tasks (such as tfjob, mpiJob, pytorchJob, sparkJob)
   - After computation is completed, data migration and cache cleaning can be carried out
- Choose data access methods based on the scheduling results of the Kubernetes scheduler (default scheduler, Volcano, YuniKorn)
   - If scheduled to ordinary nodes with shared operating system kernels, adaptively use csi plugin mode
   - If scheduled to Kata container nodes with independent operating system kernels, you can use the sidecar mode adaptively and support scalable modifications by cloud vendors

### Objective: Simplify the work of operation and maintenance and AI developers through Python SDK

- Support basic data operation
- Combine with Hugging face and Pytorch to support transparent data acceleration through pre-reading and multi-stream reading
- Support defining automated data flow operations

### Objective: Further deeply integrate the machine learning ecosystem to simplify the user experience

- Integrate with Kubeflow Pipelines to accelerate datasets in the pipeline
- Integrate with Fairing for model development and deployment in the notebook environment
- Integrate with KServe to facilitate model deployment

### Objective: Continuous security enhancement

- Minimum container permission (remove the privileged permission of FUSE Pod)
- Minimum rbac permission
- Minimal container image installation
- Continuously provide best practice documentation

### Objective: Simplicity and reliability, friendlier to users and developers

- Simplify deployment
  - Merge Dataset/Runtime controllers into one binary package
- Simplify usage
  - Support Runtimeless, Dataset as the single API entry for users to use Fluid
- Improve code quality
  - Reduce repetitive code
  - Improve test coverage
- Enhance observability
  - Provide monitoring and alerts for Datasets
- Enhance the quality of documentation
  - Organize the documentation so users can navigate it easily and find the information
  - Maintain consistency in language, style, and formatting throughout the documentation

