# image pull secrets


If the images of fluid runtime is in private docker registry, it's necessary to set image pull secret.
For related knowledge of image pull secret,     
please refer to [Pull an Image from a Private Registry
](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)

fluid supports setting image pull secrets when you deploy fluid service using helm charts

you can set the image pull secrets to `values.yaml` like the following example
```yaml
# fluid helm charts values.yaml 
# default imagePullSecrets value is empty
image:
  imagePullSecrets: []

# set values like this 
# suppose you have two image pull secret keys `test-1` and `test-2` in your cluster
image:
  imagePullSecrets: 
  - name: test-1
  - name: test-2
```

After setting `values.yaml` image pull secrets, when you have deployed the fluid service,  
you can see that controller itself already uses image pull secrets



alluxio controller yaml intercepts information
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alluxioruntime-controller
  namespace: fluid-system
spec:
  template:
    spec:
      containers:
      - image: fluidcloudnative/alluxioruntime-controller:
        name: manager
      dnsPolicy: ClusterFirst
      imagePullSecrets:
      - name: test-1
      - name: test-2
```

fluid also supports passing the image pull secrets through the controller to the runtime service that the controller pulls up

If you want to specify the image pull key for the runtime, you can do so by editing the environment variables in the deployment associated with the runtime controller that you have created.

Here is a brief YAML configuration example. When properly configured, the deployment will apply the image pull secret you specified.

```yaml
- name: IMAGE_PULL_SECRETS
  value: test-1,test-2
```

In the above YAML file, `IMAGE_PULL_SECRETS` is the name of the environment variable, and `test-1,test-2` are the image pull keys you wish to specify, separated by commas.




alluxio runtime master yaml intercepts information
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: demo-master
  namespace: test
spec:
  template:
    spec:
      containers:
      - image: fluidcloudnative/alluxio:release-2.8.1-SNAPSHOT-0433ade
        imagePullPolicy: IfNotPresent
        name: alluxio-master
      imagePullSecrets:
      - name: test1
      - name: test2

```
In conclusion, image pull secrets setting successfully


## Setting image pull secrets in JuiceFSRuntime

JuiceFSRuntime can also set image pull secrets in its spec, so different runtimes can pull from different private registries without changing the controller. `spec.imagePullSecrets` applies to the worker and fuse, and `spec.worker.imagePullSecrets` or `spec.fuse.imagePullSecrets` override it for that component. When none of them is set, the `IMAGE_PULL_SECRETS` value of the controller is used as before.

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  imagePullSecrets:
  - name: test-1
  worker:
    imagePullSecrets:
    - name: worker-registry
  fuse:
    imagePullSecrets:
    - name: fuse-registry
```

`spec.master.imagePullSecrets` and `spec.jobWorker.imagePullSecrets` have no effect for now. Changing the imagePullSecrets fields above on an existing runtime updates the worker StatefulSet and fuse DaemonSet templates on the next sync, but it doesn't restart running pods. Worker pods pick up the new secrets as they are replaced. The fuse DaemonSet uses `OnDelete`, so with the default `OnRuntimeDeleted` clean policy running fuse pods keep the old secrets until they are deleted. The same goes for the controller's `IMAGE_PULL_SECRETS` on runtimes that don't set their own. The controller owns these secrets on the worker and fuse, so edits made directly on the StatefulSet or DaemonSet are reverted.
