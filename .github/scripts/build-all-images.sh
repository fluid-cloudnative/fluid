#!/bin/bash
set -e

function get_image_tag() {
    version=$(grep "^VERSION=" ./Makefile)
    version=${version#VERSION=}

    git_sha=$(git rev-parse --short HEAD || echo "HEAD")
    export IMAGE_TAG=${version}-${git_sha}
}

function build_images() {
    images=(
        ${IMG_REPO}/dataset-controller:${IMAGE_TAG}
        ${IMG_REPO}/application-controller:${IMAGE_TAG}
        ${IMG_REPO}/alluxioruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/jindoruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/goosefsruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/juicefsruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/thinruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/efcruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/vineyardruntime-controller:${IMAGE_TAG}
        ${IMG_REPO}/fluid-csi:${IMAGE_TAG}
        ${IMG_REPO}/fluid-webhook:${IMAGE_TAG}
        ${IMG_REPO}/fluid-crd-upgrader:${IMAGE_TAG}
    )

    make docker-build-all

    for img in ${images[@]}; do
        echo "Loading image $img to kind cluster..."
        kind load docker-image $img --name ${KIND_CLUSTER}
    done
}

function cleanup_docker_caches() {
    echo ">>> System disk usage after building fluid images"
    df -h
    echo ">>> Cleaning docker caches..."
    docker system df
    docker ps
    docker container prune -f
    docker images
    docker image prune -a -f
    docker builder prune -a -f
    docker buildx prune -a -f
    echo ">>> docker caches cleaned up"
    echo ">>> System disk usage after cleaning up docker caches"
    df -h
}

get_image_tag
build_images
cleanup_docker_caches
