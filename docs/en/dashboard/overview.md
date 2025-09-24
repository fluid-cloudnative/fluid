# Dashboard

Fluid Dashboard provides a visual interface for dataset management, dataload, and runtime management, aiming to offer a better user experience for managing Fluid core resources, lowering the CLI operation threshold, and improving the management efficiency of data-intensive applications.

Currently, Fluid Dashboard offers a KubeSphere-based extension plugin. In the future, a Dashboard that supports independent operation in the native K8S environment will also be launched

## fluid kubesphere extension Main Features

- **Dataset Management**: Create, view, edit, and delete Datasets.
- **Data Loading Tasks**: Visual management of Dataload tasks, supporting multiple loading strategies.
- **Runtime Management**: Monitor the status and events of various runtimes (such as Alluxio, JuiceFS, GooseFS, etc.).
- **Multi-cluster and Namespace Support**: Switch between different clusters and namespaces, suitable for multi-tenant scenarios.

## Installation

1. Find the Fluid extension in the KubeSphere Extension Marketplace, click "Install", select the latest version, and click "Next";
2. In the extension installation tab, modify the extension configuration as needed, then click "Start Installation";
3. After installation, click "Next" to enter the cluster selection page, select the clusters to install, and click "Next" to enter the differentiated configuration page;
4. Update the differentiated configuration as needed, then start the installation and wait for completion.

## Configuration

In the extension configuration, set `enabled` to control whether to install the frontend:

```yaml
frontend:
  enabled: true
```
