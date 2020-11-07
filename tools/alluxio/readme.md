
Build the docker image for alluxio:

```bash
docker rm -f alluxio-dev-test
rm -rf /alluxio
cd /
bash build-image.sh -b branch-2.3-fuse
```