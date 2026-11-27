# Curvine access to ThinRuntime (Dev Guide)

## Prerequisites

- Kubernetes 1.18+
- Fluid installed
- A running Curvine cluster and at least one exported path

## Prepare Curvine FUSE Client Image

The ThinRuntime uses the regular Curvine runtime image, which also contains the Fluid integration scripts. Curvine publishes it as `ghcr.io/curvineio/curvine`.

1) Use the published image, or build it from source

```bash
# In Curvine repo root: compile Curvine and package it into the runtime image
make docker-build
```

`make docker-build` produces a local image named `curvine:latest`. To use it instead of the published image, tag and push it to a registry your cluster can pull from (or load it into the nodes, for example with `kind load docker-image`), and set `spec.fuse.image` and `spec.fuse.imageTag` in the ThinRuntimeProfile below to that image:

```bash
docker tag curvine:latest <your-registry>/curvine:<tag>
docker push <your-registry>/curvine:<tag>
```

2) How the image works

- Started with `/entrypoint.sh fluid-thin-runtime` (or when Fluid mounts its config), it hands over to `/fluid-entrypoint.sh`
- Reads Fluid-generated config from /etc/fluid/config/config.json
- Parses Dataset/ThinRuntime/ThinRuntimeProfile into Curvine config via `/app/curvine/config-parse.py`
- Launches curvine-fuse and keeps the process running

## Create ThinRuntimeProfile

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: curvine
spec:
  fileSystemType: fuse
  fuse:
    image: ghcr.io/curvineio/curvine
    imageTag: latest
    imagePullPolicy: IfNotPresent
    command:
    - /entrypoint.sh
    args:
    - fluid-thin-runtime
```

## Create Dataset and ThinRuntime

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: curvine-demo
spec:
  mounts:
    - mountPoint: curvine:///data
      name: curvine
      options:
        master-endpoints: "<CURVINE_MASTER_IP:PORT>"
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: curvine-demo
spec:
  profileName: curvine
```

## Run Application Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      command: ["bash"]
      args: ["-c", "sleep 9999"]
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: curvine-demo
```

## Notes

- master-endpoints should be reachable from the FUSE Pod
- Ensure the Curvine FUSE binary exists in the image and glibc versions match
- For a complete example, see `curvine-docker/fluid/thin-runtime/curvine-thinruntime.yaml` in the Curvine repo
