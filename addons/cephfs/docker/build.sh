#!/usr/bin/env bash
set +x

docker build . --network=host -f Dockerfile -t baowj/cephfs:v0.1

docker push baowj/cephfs:v0.1
