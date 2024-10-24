#!/usr/bin/env bash
set +x

docker build . --network=host -f Dockerfile -t fluidcloudnative/glusterfs:v0.1

docker push fluidcloudnative/glusterfs:v0.1