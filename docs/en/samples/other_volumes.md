# DEMO - Mount Other Volumes into the Runtime Pods

An AlluxioRuntime can mount extra Kubernetes volumes, such as a Secret, into its pods: declare them in `spec.volumes`, then mount them in a component with `volumeMounts` (here `spec.master.volumeMounts`).

1. Create a Secret

```
kubectl create secret generic fluid-secret --from-literal=username=test --from-literal=password=test
```

2. Check the Secret

```
kubectl get secret fluid-secret -oyaml
apiVersion: v1
data:
  password: dGVzdA==
  username: dGVzdA==
kind: Secret
metadata:
  creationTimestamp: "2022-05-13T07:49:00Z"
  name: fluid-secret
  namespace: default
  resourceVersion: "1059299"
  uid: 92e310f0-b1df-4c9c-9d58-51497eb6c89a
type: Opaque
```

3. Create the Dataset and the AlluxioRuntime

```yaml
cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
      name: hbase
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  master:
    volumeMounts:
      - name: secret-volume
        readOnly: true
        mountPath: "/etc/secret-volume"
  volumes:
    - name: secret-volume
      secret:
        secretName: fluid-secret
EOF
```

4. Log in to the master pod and check the files

```bash
# kubectl exec -it hbase-master-0 bash
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [POD] -- [COMMAND] instead.
Defaulted container "alluxio-master" out of: alluxio-master, alluxio-job-master
# mount |grep secret-volume
tmpfs on /etc/secret-volume type tmpfs (ro,relatime,size=28147972k)
# cat /etc/secret-volume/username
test
```
