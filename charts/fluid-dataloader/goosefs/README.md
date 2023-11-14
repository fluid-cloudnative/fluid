# fluid-dataloader

## Prerequisite
- Dataset deployed
- GooseFS Runtime deployed
- Dataset mountPoint mounted
- Dataset-related PV, PVC created

## Install
1. get dataset-related PVC name
```shell script
$ kubectl get pvc
NAME         STATUS   VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
<pvc-name>   Bound    <pv-name>    100Gi      RWX                           4h5m
```
Say `<pvc-name>` is the name of your dataset-related PVC, usually it's the same name as your dataset.

2. get num of GooseFS workers
```shell script
kubectl get pod -l release=<dataset-name> | grep -c "worker"
```

3. Install fluid-dataloader

```shell script
helm install \
  --set dataloader.numWorker=<num-of-workers> \
  --set dataloader.threads=2 \
  <pvc-name>-load charts/fluid-dataloader
```

You will see something like this:
```
helm install hbase-load charts/fluid-dataloader/
NAME: <pvc-name>-load
LAST DEPLOYED: Fri Jul 31 19:52:11 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Some dataloader jobs will be launched. You will see multiple jobs running on different nodes:
```shell script
kubectl get pod -o wide -l role=goosefs-dataloader
```

Once some job completes, you can check time consumed during data prefetch:
```shell script
kubectl logs <pvc-name>-loader-xxxxx
```
and see something like this:
```
THREADS=2
DATAPATH=/data/*
python multithread_read_benchmark.py --threads=2 --path=/data/*
/data/* contains 15 items
/data/* processing 15 items with 2 threads uses 32.6712441444s, avg 0.459119338513/s, avg 8743748.5924B/s, avg 8.33868846169MiB/s
```

Now then, all data should be cached, reinstall it:
```shell script
helm del <pvc-name>

helm install \
  --set dataloader.numWorker=<num-of-workers> \
  --set dataloader.threads=2 \
  <pvc-name>-load charts/fluid-dataloader
```

check again, and this time should be much faster:
```shell script
kubectl logs <pvc-name>-loader-yyyyy
```
```
THREADS=2
DATAPATH=/data/*
python multithread_read_benchmark.py --threads=2 --path=/data/*
/data/* contains 15 items
/data/* processing 15 items with 2 threads uses 0.308158159256s, avg 48.6763032211/s, avg 927021194.862B/s, avg 884.076304304MiB/s
```

## Uninstall
```
helm del <pvc-name>
```
