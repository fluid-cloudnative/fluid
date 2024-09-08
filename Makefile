# Version and Git information
VERSION := v1.0.3
BUILD_DATE := $(shell date -u +'%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_TAG := $(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE := $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
GIT_SHA := $(shell git rev-parse --short HEAD || echo "HEAD")
GIT_VERSION := ${VERSION}-${GIT_SHA}
PACKAGE := github.com/fluid-cloudnative/fluid

# Go and build settings
GO_MODULE ?= off
GC_FLAGS ?= -gcflags="all=-N -l"
LOCAL_FLAGS ?= -gcflags=-l
CGO_ENABLED ?= 0
GOOS ?= linux
GOBIN := $(shell if [ -z "$(shell go env GOBIN)" ]; then echo "$(shell go env GOPATH)/bin"; else echo "$(shell go env GOBIN)"; fi)

# Architecture detection
UNAME := $(shell uname -m)
ifeq ($(UNAME), aarch64)
    ARCH := arm64
else
    ARCH := amd64
endif

# Docker settings
DOCKER_PLATFORM ?= linux/amd64,linux/arm64
NO_CACHE ?=
DOCKER_NO_CACHE_OPTION := $(if $(filter true,${NO_CACHE}),--no-cache,)

# Image repository and component images
IMG_REPO ?= fluidcloudnative
DATASET_CONTROLLER_IMG ?= ${IMG_REPO}/dataset-controller
APPLICATION_CONTROLLER_IMG ?= ${IMG_REPO}/application-controller
ALLUXIORUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/alluxioruntime-controller
JINDORUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/jindoruntime-controller
GOOSEFSRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/goosefsruntime-controller
JUICEFSRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/juicefsruntime-controller
THINRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/thinruntime-controller
EFCRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/efcruntime-controller
VINEYARDRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/vineyardruntime-controller
CSI_IMG ?= ${IMG_REPO}/fluid-csi
LOADER_IMG ?= ${IMG_REPO}/fluid-dataloader
INIT_USERS_IMG ?= ${IMG_REPO}/init-users
WEBHOOK_IMG ?= ${IMG_REPO}/fluid-webhook
CRD_UPGRADER_IMG ?= ${IMG_REPO}/fluid-crd-upgrader

# Dockerfile paths
DATASET_DOCKERFILE ?= docker/Dockerfile.dataset
APPLICATION_DOCKERFILE ?= docker/Dockerfile.application
ALLUXIORUNTIME_DOCKERFILE ?= docker/Dockerfile.alluxioruntime
JINDORUNTIME_DOCKERFILE ?= docker/Dockerfile.jindoruntime
GOOSEFSRUNTIME_DOCKERFILE ?= docker/Dockerfile.goosefsruntime
JUICEFSRUNTIME_DOCKERFILE ?= docker/Dockerfile.juicefsruntime
THINRUNTIME_DOCKERFILE ?= docker/Dockerfile.thinruntime
EFCRUNTIME_DOCKERFILE ?= docker/Dockerfile.efcruntime
VINEYARDRUNTIME_DOCKERFILE ?= docker/Dockerfile.vineyardruntime
CSI_DOCKERFILE ?= docker/Dockerfile.csi
INIT_USERS_DOCKERFILE ?= charts/alluxio/docker/init-users
WEBHOOK_DOCKERFILE ?= docker/Dockerfile.webhook
CRD_UPGRADER_DOCKERFILE ?= docker/Dockerfile.crds

# Binary paths
CSI_BINARY ?= bin/fluid-csi
DATASET_BINARY ?= bin/dataset-controller
APPLICATION_BINARY ?= bin/fluidapp-controller
ALLUXIORUNTIME_BINARY ?= bin/alluxioruntime-controller
JINDORUNTIME_BINARY ?= bin/jindoruntime-controller
GOOSEFSRUNTIME_BINARY ?= bin/goosefsruntime-controller
JUICEFSRUNTIME_BINARY ?= bin/juicefsruntime-controller
THINRUNTIME_BINARY ?= bin/thinruntime-controller
EFCRUNTIME_BINARY ?= bin/efcruntime-controller
VINEYARDRUNTIME_BINARY ?= bin/vineyardruntime-controller
WEBHOOK_BINARY ?= bin/fluid-webhook

# Miscellaneous
HELM_VERSION ?= helm-v3.11.3
CRD_OPTIONS ?= "crd"

# Build binaries
BINARY_BUILD := dataset-controller-build
BINARY_BUILD += application-controller-build
BINARY_BUILD += alluxioruntime-controller-build
BINARY_BUILD += jindoruntime-controller-build
BINARY_BUILD += juicefsruntime-controller-build
BINARY_BUILD += thinruntime-controller-build
BINARY_BUILD += efcruntime-controller-build
BINARY_BUILD += vineyardruntime-controller-build
BINARY_BUILD += csi-build
BINARY_BUILD += webhook-build

# Build docker images
DOCKER_BUILD := docker-build-dataset-controller
DOCKER_BUILD += docker-build-application-controller
DOCKER_BUILD += docker-build-alluxioruntime-controller
DOCKER_BUILD += docker-build-jindoruntime-controller
DOCKER_BUILD += docker-build-goosefsruntime-controller
DOCKER_BUILD += docker-build-csi
DOCKER_BUILD += docker-build-webhook
DOCKER_BUILD += docker-build-juicefsruntime-controller
DOCKER_BUILD += docker-build-thinruntime-controller
DOCKER_BUILD += docker-build-efcruntime-controller
DOCKER_BUILD += docker-build-vineyardruntime-controller
DOCKER_BUILD += docker-build-init-users
DOCKER_BUILD += docker-build-crd-upgrader
DOCKER_BUILD += docker-build-loader

# Push docker images
DOCKER_PUSH := docker-push-dataset-controller
DOCKER_PUSH += docker-push-application-controller
DOCKER_PUSH += docker-push-alluxioruntime-controller
DOCKER_PUSH += docker-push-jindoruntime-controller
DOCKER_PUSH += docker-push-csi
DOCKER_PUSH += docker-push-webhook
DOCKER_PUSH += docker-push-goosefsruntime-controller
DOCKER_PUSH += docker-push-juicefsruntime-controller
DOCKER_PUSH += docker-push-thinruntime-controller
DOCKER_PUSH += docker-push-efcruntime-controller
DOCKER_PUSH += docker-push-vineyardruntime-controller
# Not need to push init-users image by default
# DOCKER_PUSH += docker-push-init-users
DOCKER_PUSH += docker-push-crd-upgrader

# Buildx and push docker images
DOCKER_BUILDX_PUSH := docker-buildx-push-dataset-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-application-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-alluxioruntime-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-jindoruntime-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-goosefsruntime-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-csi
DOCKER_BUILDX_PUSH += docker-buildx-push-webhook
DOCKER_BUILDX_PUSH += docker-buildx-push-juicefsruntime-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-thinruntime-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-efcruntime-controller
DOCKER_BUILDX_PUSH += docker-buildx-push-init-users
DOCKER_BUILDX_PUSH += docker-buildx-push-crd-upgrader

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE} \
  -extldflags "-static"

all: build

# Run tests
test: generate fmt vet
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go list ./... | grep -v controller | grep -v e2etest | FLUID_UNIT_TEST=true xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# used in CI and simply ignore controller tests which need k8s now.
# maybe incompatible if more end to end tests are added.
unit-test: generate fmt vet
	GO111MODULE=${GO_MODULE} go list ./... | grep -v controller | grep -v e2etest | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# Make code, artifacts, dependencies, and CRDs fresh.
pre-setup: generate fmt vet update-crd gen-openapi

# Generate code
generate: controller-gen
	GO111MODULE=${GO_MODULE} $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Run go fmt against code
fmt:
	GO111MODULE=${GO_MODULE} go fmt ./...

# Run go vet against code
vet:
	GO111MODULE=${GO_MODULE} go list ./... | grep -v "vendor" | xargs go vet

# Update fluid crds
update-crd: manifests
	cp config/crd/bases/* charts/fluid/fluid/crds

gen-openapi:
	./hack/gen-openapi.sh

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	GO111MODULE=${GO_MODULE} $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

gen-sdk:
	./hack/sdk/gen-sdk.sh

update-api-doc:
	bash tools/api-doc-gen/generate_api_doc.sh && mv tools/api-doc-gen/api_doc.md docs/zh/dev/api_doc.md && cp docs/zh/dev/api_doc.md docs/en/dev/api_doc.md

# Build binary
build: pre-setup ${BINARY_BUILD}

csi-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${CSI_BINARY} -ldflags '${LDFLAGS}' cmd/csi/main.go

dataset-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${DATASET_BINARY} -ldflags '${LDFLAGS}' cmd/dataset/main.go

alluxioruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${ALLUXIORUNTIME_BINARY} -ldflags '${LDFLAGS}' cmd/alluxio/main.go

jindoruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${JINDORUNTIME_BINARY} -ldflags '${LDFLAGS}' cmd/jindo/main.go

goosefsruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${GOOSEFSRUNTIME_BINARY} -ldflags '${LDFLAGS}' cmd/goosefs/main.go

juicefsruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${JUICEFSRUNTIME_BINARY} -ldflags '-s -w ${LDFLAGS}' cmd/juicefs/main.go

thinruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${THINRUNTIME_BINARY} -ldflags '-s -w ${LDFLAGS}' cmd/thin/main.go

vineyardruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${VINEYARDRUNTIME_BINARY} -ldflags '-s -w ${LDFLAGS}' cmd/vineyard/main.go

efcruntime-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${EFCRUNTIME_BINARY} -ldflags '${LDFLAGS}' cmd/efc/main.go

webhook-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${WEBHOOK_BINARY} -ldflags '${LDFLAGS}' cmd/webhook/main.go

application-controller-build:
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o ${APPLICATION_BINARY} -ldflags '${LDFLAGS}' cmd/fluidapp/main.go

# Build the docker image
docker-build-dataset-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${DATASET_DOCKERFILE} -t ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-application-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${APPLICATION_DOCKERFILE} -t ${APPLICATION_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-alluxioruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${ALLUXIORUNTIME_DOCKERFILE} -t ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-jindoruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${JINDORUNTIME_DOCKERFILE} -t ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-goosefsruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${GOOSEFSRUNTIME_DOCKERFILE} -t ${GOOSEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-juicefsruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${JUICEFSRUNTIME_DOCKERFILE} -t ${JUICEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-thinruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${THINRUNTIME_DOCKERFILE} -t ${THINRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-efcruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${EFCRUNTIME_DOCKERFILE} -t ${EFCRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-vineyardruntime-controller:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} --build-arg HELM_VERSION=${HELM_VERSION} . -f ${VINEYARDRUNTIME_DOCKERFILE} -t ${VINEYARDRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-csi:
	docker build ${DOCKER_NO_CACHE_OPTION} . -f ${CSI_DOCKERFILE} -t ${CSI_IMG}:${GIT_VERSION}

docker-build-init-users:
	docker build ${DOCKER_NO_CACHE_OPTION} ${INIT_USERS_DOCKERFILE} -t ${INIT_USERS_IMG}:${VERSION}

docker-build-webhook:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} . -f ${WEBHOOK_DOCKERFILE} -t ${WEBHOOK_IMG}:${GIT_VERSION}

docker-build-crd-upgrader:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} . -f ${CRD_UPGRADER_DOCKERFILE} -t ${CRD_UPGRADER_IMG}:${GIT_VERSION}

docker-build-loader:
	docker build ${DOCKER_NO_CACHE_OPTION} --build-arg TARGETARCH=${ARCH} . -f ${LOADER_DOCKERFILE} -t ${LOADER_IMG}:${GIT_VERSION}

# Push the docker image
docker-push-dataset-controller: docker-build-dataset-controller
	docker push ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-application-controller: docker-build-application-controller
	docker push ${APPLICATION_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-alluxioruntime-controller: docker-build-alluxioruntime-controller
	docker push ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-jindoruntime-controller: docker-build-jindoruntime-controller
	docker push ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-goosefsruntime-controller: docker-build-goosefsruntime-controller
	docker push ${GOOSEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-juicefsruntime-controller: docker-build-juicefsruntime-controller
	docker push ${JUICEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-thinruntime-controller: docker-build-thinruntime-controller
	docker push ${THINRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-efcruntime-controller: docker-build-efcruntime-controller
	docker push ${EFCRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-vineyardruntime-controller: docker-build-vineyardruntime-controller
	docker push ${VINEYARDRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-csi: docker-build-csi
	docker push ${CSI_IMG}:${GIT_VERSION}

docker-push-loader: docker-build-loader
	docker push ${LOADER_IMG}:${GIT_VERSION}

docker-push-init-users: docker-build-init-users
	docker push ${INIT_USERS_IMG}:${VERSION}

docker-push-webhook: docker-build-webhook
	docker push ${WEBHOOK_IMG}:${GIT_VERSION}

docker-push-crd-upgrader: docker-build-crd-upgrader
	docker push ${CRD_UPGRADER_IMG}:${GIT_VERSION}

# Buildx and push the docker image
docker-buildx-push-dataset-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${DATASET_DOCKERFILE} -t ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-application-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${APPLICATION_DOCKERFILE} -t ${APPLICATION_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-alluxioruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${ALLUXIORUNTIME_DOCKERFILE} -t ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-jindoruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${JINDORUNTIME_DOCKERFILE} -t ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-goosefsruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${GOOSEFSRUNTIME_DOCKERFILE} -t ${GOOSEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-juicefsruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${JUICEFSRUNTIME_DOCKERFILE} -t ${JUICEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-thinruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${THINRUNTIME_DOCKERFILE} -t ${THINRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-efcruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${EFCRUNTIME_DOCKERFILE} -t ${EFCRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-vineyardruntime-controller:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${VINEYARDRUNTIME_DOCKERFILE} -t ${VINEYARDRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-csi: generate fmt vet
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${CSI_DOCKERFILE} -t ${CSI_IMG}:${GIT_VERSION}

docker-buildx-push-init-users:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} ${INIT_USERS_DOCKERFILE} -t ${INIT_USERS_IMG}:${VERSION}

docker-buildx-push-webhook:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${WEBHOOK_DOCKERFILE} -t ${WEBHOOK_IMG}:${GIT_VERSION}

docker-buildx-push-crd-upgrader:
	docker buildx build --push --platform ${DOCKER_PLATFORM} ${DOCKER_NO_CACHE_OPTION} . -f ${CRD_UPGRADER_DOCKERFILE} -t ${CRD_UPGRADER_IMG}:${GIT_VERSION}

docker-build-all: pre-setup ${DOCKER_BUILD}
docker-push-all: pre-setup ${DOCKER_PUSH}
docker-buildx-all-push: pre-setup ${DOCKER_BUILDX_PUSH}


# find or download controller-gen
# download controller-gen if necessary
# controller-gen@v0.14.0 comply with k8s.io/api v0.29.x
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	export GO111MODULE=on ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
