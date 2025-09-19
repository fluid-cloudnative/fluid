# Curvine access to ThinRuntime (Dev Guide)

## Prerequisites

- Kubernetes 1.18+
- Fluid installed
- A running Curvine cluster and at least one exported path

## Prepare Curvine FUSE Client Image

This guide uses the Curvine ThinRuntime build assets.

1) Build Curvine binaries

```bash
# In Curvine repo root
make all
# Expected output includes build/dist/lib/curvine-fuse
```

2) Build ThinRuntime image

```bash
# In curvine repo
cd curvine-docker/fluid/thin-runtime
./build-image.sh
```

The resulting image should be something like:

- fluid-cloudnative/curvine-thinruntime:v1.0.0

3) How the image works

- Reads Fluid-generated config from /etc/fluid/config/config.json
- Parses Dataset/ThinRuntime/ThinRuntimeProfile into Curvine config via fluid-config-parse.py
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
    image: fluid-cloudnative/curvine-thinruntime
    imageTag: v1.0.0
    imagePullPolicy: IfNotPresent
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
- For more details, see Curvine ThinRuntime README in Curvine repo
