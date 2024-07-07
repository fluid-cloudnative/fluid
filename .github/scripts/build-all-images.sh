#!/bin/bash
set -e

function get_image_tag() {
    version=$(grep "^VERSION=" ./Makefile)
    version=${version#VERSION=}

    git_sha=$(git rev-parse --short HEAD || echo "HEAD")
    export IMAGE_TAG=${version}-${git_sha}
}

get_image_tag

make docker-build-all

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

for img in ${images[@]}; do
    echo "Loading image $img to kind cluster..."
    kind load docker-image $img --name ${KIND_CLUSTER}
done