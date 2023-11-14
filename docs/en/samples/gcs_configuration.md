## DEMO - The special configurations required for Fluid access Google Cloud Storage (GCS)

If `Google Cloud Storage (GCS)` is selected as the underlying storage system of Alluxio, additional configuration is required for Alluxio to properly access the mounted GCS storage system.

This document shows how to do the special configuration required for Alluxio in Fluid in a declarative manner. For more information, see [Configuring Alluxio on Google Cloud Storage (GCS)](https://docs.alluxio.io/os/user/stable/en/ufs/GCS.html).

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- The [GCS bucket](https://cloud.google.com/storage/docs/creating-buckets)
- [Google Application Credentials](https://cloud.google.com/docs/authentication/getting-started) that has permission to access the bucket

Please refer to the [Fluid installation documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) to complete the installation.

## Run the example

In this example, we will use `GCS UFS version 2` which uses `Google Application Credentials` to access GCS. It is the recommended way to access GCS.

Accessing GCS in Alluxio is a bit different from other object storage systems like S3. You have to mount the `Google Application Credentials` which is a `json` file into your Alluxio master and workers, and provide the path to the json file in the Alluxio properties.

**Create Secret**

Run the following command to create a secret called `gcscreds` with key `gcloud-application-credentials.json` and value of the json file.

```sh
kubectl create secret generic gcscreds \
    --from-file=gcloud-application-credentials.json=<name of your google application credentials>.json
```

**Create AlluxioRuntime Resource Object**

Here you create your Alluxio runtime and mount the above secret into master and workers, the path in both master and worker must be the same.

```yaml
$ cat << EOF > runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: gcs-data
spec:
  ...
  volumes:
    - name: secret-volume
      secret:
        secretName: gcscreds # secret name you created
  master:
    volumeMounts:
      - name: secret-volume
        readOnly: true
        mountPath: "/var/secrets"
  worker:
    volumeMounts:
      - name: secret-volume
        readOnly: true
        mountPath: "/var/secrets"
EOF
```

Create the runtime

```
$ kubectl create -f runtime.yaml
```

**Create Dataset Resource Object**

Here you specify your GCS bucket and directory in the `mountPoint`, and provide the path to your `google application credentials` (the path you mounted in the runtime previously) in the `options`.

```yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: gcs-data
spec:
  mounts:
    - mountPoint: gs://<bucket-name>/<path-to-data>
      name: gcs
      options:
        fs.gcs.credential.path: /var/secrets/gcloud-application-credentials.json
  accessModes:
    - ReadOnlyMany
EOF
```

Create your dataset

```
$ kubectl create -f dataset.yaml
```

The bucket specified in 'dataset.yaml' will be mounted to the '/gcs' directory in Alluxio.