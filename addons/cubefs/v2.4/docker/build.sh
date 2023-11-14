#!/usr/bin/env bash
set +x

docker build . --network=host -f Dockerfile -t fluidcloudnative/cubefs_v2.4:v0.1

docker push fluidcloudnative/cubefs_v2.4:v0.1