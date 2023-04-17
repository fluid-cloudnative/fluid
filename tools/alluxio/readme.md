
Build the docker image for alluxio:

```bash
docker rm -f alluxio-dev-test
rm -rf /alluxio
cd /
bash build-image.sh -b branch-2.3-fuse
```
You can run this script to customize your own image with the following parameters:
```bash
-h, --help
    list the help messages.
-b, --branch branch
    Set the git branch. If you don't assign it, the default branch is "branch-2.3-fuse".
-t, --tag tag
    Set the git tag. If you don't assign it, the default tag is empty.
-c, --commit
    Set the commit_id. If you don't assign it, the default commit_id is empty.
-a, --alluxio_image_name alluxio_image_name
    Set the alluxio image name. If you don't assign it, the default alluxio image name is "registry.aliyuncs.com/alluxio/alluxio".
-f, --alluxio_fuse_image_name alluxio_fuse_image_name
    Set the alluxio fuse image name. If you don't assign it, the default alluxio fuse image name is "registry.aliyuncs.com/alluxio/alluxio-fuse".
```