#!/usr/bin/env bash
set +x

docker build . --network=host -f Dockerfile -t fluidcloudnative/ossfs:v1.0

docker push fluidcloudnative/ossfs:v1.0
