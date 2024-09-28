#!/bin/bash
set -e

function get_image_tag() {
    version=$(grep "^VERSION := " ./Makefile)
    version=${version#VERSION := }

    git_sha=$(git rev-parse --short HEAD || echo "HEAD")
    export IMAGE_TAG=${version}-${git_sha}
}

function deploy_fluid() {
    echo "Replacing image tags in values.yaml with $IMAGE_TAG"
    sed -i -E "s/version: &defaultVersion v[0-9]\.[0-9]\.[0-9]-[a-z0-9]+$/version: \&defaultVersion $IMAGE_TAG/g" charts/fluid/fluid/values.yaml
    kubectl create ns fluid-system
    helm install --create-namespace --set runtime.jindo.smartdata.imagePrefix=registry.cn-hongkong.aliyuncs.com/jindofs --set runtime.jindo.fuse.imagePrefix=registry.cn-hongkong.aliyuncs.com/jindofs fluid charts/fluid/fluid
}

function main() {
    get_image_tag
    if [[ -z "$IMAGE_TAG" ]];then
        echo "Failed to get image tag, exiting..."
        exit 1
    fi
    
    deploy_fluid
}

main
