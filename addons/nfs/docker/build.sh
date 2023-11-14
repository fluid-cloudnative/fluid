#!/usr/bin/env bash
set +x

docker build . --network=host -f Dockerfile -t fluidcloudnative/nfs:v0.1

docker push fluidcloudnative/nfs:v0.1