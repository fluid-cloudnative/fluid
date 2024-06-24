#!/usr/bin/env bash
set +x

docker build . --network=host -f Dockerfile -t fluidcloudnative/dynamic-mount:base

docker push fluidcloudnative/dynamic-mount:base
