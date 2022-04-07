# fluid-databackup

## Prerequisite
- Dataset deployed
- GooseFS Runtime deployed
- Dataset mountPoint mounted
- Dataset-related PV, PVC created

## Install
1. Install fluid-databackup

```shell script
helm install charts/fluid-databackup
```

You will see something like this:
```
helm install charts/fluid-databackup
NAME: test
LAST DEPLOYED: Fri Jan 15 09:18:02 2021
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

one datbackup pod will be launched. You will see one pod running on the node:
```shell script
kubectl get pods <datasetname>-databackup-pod -o wide
```

Once the pod completes, you can check filed backuped:
```shell script
$ ls
hbase-default.yaml  metadata-backup-hbase-default.gz
```

## Uninstall
```
helm del test
```
