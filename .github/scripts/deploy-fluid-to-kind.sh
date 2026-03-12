#! /bin/bash

get_image_tag() {
    local version=""
    version=$(grep "^VERSION := " ./Makefile)
    version="${version#VERSION := }"

    local git_sha=""
    git_sha=$(git rev-parse --short HEAD || echo "HEAD")
    export IMAGE_TAG="${version}-${git_sha}"
}

deploy_fluid() {
    echo "Replacing image tags in values.yaml with ${IMAGE_TAG}"
    sed -i -E "s/version: &defaultVersion .+$/version: \&defaultVersion ${IMAGE_TAG}/g" charts/fluid/fluid/values.yaml
    kubectl create ns fluid-system || true
    helm upgrade --install --namespace fluid-system --create-namespace --set runtime.jindo.smartdata.imagePrefix=registry-cn-hongkong.ack.aliyuncs.com/acs --set runtime.jindo.fuse.imagePrefix=registry-cn-hongkong.ack.aliyuncs.com/acs fluid charts/fluid/fluid
}

main() {
    get_image_tag
    if [[ -z "${IMAGE_TAG}" ]]; then
        echo "Failed to get image tag, exiting..."
        exit 1
    fi
    
    deploy_fluid
}

main
