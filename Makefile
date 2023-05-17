# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

# The Image URL to use in docker build and push
# IMG_REPO ?= registry.aliyuncs.com/fluid
IMG_REPO ?= fluidcloudnative
DATASET_CONTROLLER_IMG ?= ${IMG_REPO}/dataset-controller
APPLICATION_CONTROLLER_IMG ?= ${IMG_REPO}/application-controller
ALLUXIORUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/alluxioruntime-controller
JINDORUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/jindoruntime-controller
GOOSEFSRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/goosefsruntime-controller
JUICEFSRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/juicefsruntime-controller
THINRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/thinruntime-controller
EFCRUNTIME_CONTROLLER_IMG ?= ${IMG_REPO}/efcruntime-controller
CSI_IMG ?= ${IMG_REPO}/fluid-csi
LOADER_IMG ?= ${IMG_REPO}/fluid-dataloader
INIT_USERS_IMG ?= ${IMG_REPO}/init-users
WEBHOOK_IMG ?= ${IMG_REPO}/fluid-webhook
CRD_UPGRADER_IMG?= ${IMG_REPO}/fluid-crd-upgrader
GO_MODULE ?= off
GC_FLAGS ?= -gcflags="all=-N -l"
ARCH ?= amd64

LOCAL_FLAGS ?= -gcflags=-l
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

UNAME := $(shell uname -m)
ifeq ($(UNAME), aarch64)
    ARCH := arm64
else
    ARCH := amd64
endif

CURRENT_DIR=$(shell pwd)
VERSION=v0.9.0
BUILD_DATE=$(shell date -u +'%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
GIT_SHA=$(shell git rev-parse --short HEAD || echo "HEAD")
GIT_VERSION=${VERSION}-${GIT_SHA}
PACKAGE=github.com/fluid-cloudnative/fluid

# Build binaries
BINARY_BUILD := dataset-controller-build
BINARY_BUILD += application-controller-build
BINARY_BUILD += alluxioruntime-controller-build
BINARY_BUILD += jindoruntime-controller-build
BINARY_BUILD += juicefsruntime-controller-build
BINARY_BUILD += thinruntime-controller-build
BINARY_BUILD += efcruntime-controller-build
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
DOCKER_BUILD += docker-build-init-users
DOCKER_BUILD += docker-build-crd-upgrader

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
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go list ./... | grep -v controller | grep -v e2etest | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# used in CI and simply ignore controller tests which need k8s now.
# maybe incompatible if more end to end tests are added.
unit-test: generate fmt vet
	GO111MODULE=${GO_MODULE} go list ./... | grep -v controller | grep -v e2etest | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# Build binary

build: ${BINARY_BUILD}

csi-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/fluid-csi -ldflags '${LDFLAGS}' cmd/csi/main.go

dataset-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/dataset-controller -ldflags '${LDFLAGS}' cmd/dataset/main.go

alluxioruntime-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/alluxioruntime-controller -ldflags '${LDFLAGS}' cmd/alluxio/main.go

jindoruntime-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/jindoruntime-controller -ldflags '${LDFLAGS}' cmd/jindo/main.go

goosefsruntime-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/goosefsruntime-controller -ldflags '${LDFLAGS}' cmd/goosefs/main.go

juicefsruntime-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/juicefsruntime-controller -ldflags '-s -w ${LDFLAGS}' cmd/juicefs/main.go

thinruntime-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/thinruntime-controller -ldflags '-s -w ${LDFLAGS}' cmd/thin/main.go

efcruntime-controller-build: generate gen-openapi fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/efcruntime-controller -ldflags '${LDFLAGS}' cmd/efc/main.go

webhook-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/fluid-webhook -ldflags '${LDFLAGS}' cmd/webhook/main.go

application-controller-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go build ${GC_FLAGS} -a -o bin/fluidapp-controller -ldflags '${LDFLAGS}' cmd/fluidapp/main.go

# Debug against the configured Kubernetes cluster in ~/.kube/config, add debug
debug: generate fmt vet manifests
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  dlv debug --headless --listen ":12345" --log --api-version=2 cmd/controller/main.go

# Debug against the configured Kubernetes cluster in ~/.kube/config, add debug
debug-csi: generate fmt vet manifests
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  dlv debug --headless --listen ":12346" --log --api-version=2 cmd/csi/main.go -- --nodeid=cn-hongkong.172.31.136.194 --endpoint=unix://var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=${GO_MODULE}  go run cmd/controller/main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	GO111MODULE=${GO_MODULE} $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against codecsi-node-driver-registrar
fmt:
	GO111MODULE=${GO_MODULE} go fmt ./...

# Run go vet against code
vet:
	GO111MODULE=${GO_MODULE} go list ./... | grep -v "vendor" | xargs go vet

# Generate code
generate: controller-gen
	GO111MODULE=${GO_MODULE} $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Update fluid helm chart
update-crd: manifests
	cp config/crd/bases/* charts/fluid/fluid/crds

update-api-doc:
	bash tools/api-doc-gen/generate_api_doc.sh && mv tools/api-doc-gen/api_doc.md docs/zh/dev/api_doc.md && cp docs/zh/dev/api_doc.md docs/en/dev/api_doc.md

# Build the docker image
docker-build-dataset-controller: generate gen-openapi fmt vet
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.dataset -t ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-application-controller: generate fmt vet
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.application -t ${APPLICATION_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-alluxioruntime-controller: generate gen-openapi fmt vet
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.alluxioruntime -t ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-jindoruntime-controller: generate gen-openapi fmt vet
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.jindoruntime -t ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-goosefsruntime-controller: generate gen-openapi fmt vet
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.goosefsruntime -t ${GOOSEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-juicefsruntime-controller: generate gen-openapi fmt vet juicefsruntime-controller-build
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.juicefsruntime -t ${JUICEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-thinruntime-controller: generate gen-openapi fmt vet thinruntime-controller-build
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.thinruntime -t ${THINRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-efcruntime-controller: generate gen-openapi fmt vet
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.efcruntime -t ${EFCRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-csi: generate fmt vet
	docker build --no-cache . -f docker/Dockerfile.csi -t ${CSI_IMG}:${GIT_VERSION}

docker-build-loader:
	docker build --no-cache charts/fluid-dataloader/docker/loader -t ${LOADER_IMG}

docker-build-init-users:
	docker build --no-cache charts/alluxio/docker/init-users -t ${INIT_USERS_IMG}:${VERSION}

docker-build-webhook:
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.webhook -t ${WEBHOOK_IMG}:${GIT_VERSION}

docker-build-crd-upgrader:
	docker build --no-cache --build-arg TARGETARCH=${ARCH} . -f docker/Dockerfile.crds -t ${CRD_UPGRADER_IMG}:${GIT_VERSION}

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

docker-push-csi: docker-build-csi
	docker push ${CSI_IMG}:${GIT_VERSION}

docker-push-loader: docker-build-loader
	docker push ${LOADER_IMG}

docker-push-init-users: docker-build-init-users
	docker push ${INIT_USERS_IMG}:${VERSION}

docker-push-webhook: docker-build-webhook
	docker push ${WEBHOOK_IMG}:${GIT_VERSION}

docker-push-crd-upgrader: docker-build-crd-upgrader
	docker push ${CRD_UPGRADER_IMG}:${GIT_VERSION}

# Buildx and push the docker image
docker-buildx-push-dataset-controller: generate gen-openapi fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.dataset -t ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-application-controller: generate fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.application -t ${APPLICATION_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-alluxioruntime-controller: generate gen-openapi fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.alluxioruntime -t ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-jindoruntime-controller: generate gen-openapi fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.jindoruntime -t ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-goosefsruntime-controller: generate gen-openapi fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.goosefsruntime -t ${GOOSEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-juicefsruntime-controller: generate gen-openapi fmt vet juicefsruntime-controller-build
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.juicefsruntime -t ${JUICEFSRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-thinruntime-controller: generate gen-openapi fmt vet thinruntime-controller-build
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.thinruntime -t ${THINRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-efcruntime-controller: generate gen-openapi fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.efcruntime -t ${EFCRUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-buildx-push-csi: generate fmt vet
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.csi -t ${CSI_IMG}:${GIT_VERSION}

docker-buildx-push-init-users:
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache charts/alluxio/docker/init-users -t ${INIT_USERS_IMG}:${VERSION}

docker-buildx-push-webhook:
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.webhook -t ${WEBHOOK_IMG}:${GIT_VERSION}

docker-buildx-push-crd-upgrader:
	docker buildx build --push --platform linux/amd64,linux/arm64 --no-cache . -f docker/Dockerfile.crds -t ${CRD_UPGRADER_IMG}:${GIT_VERSION}

docker-build-all: ${DOCKER_BUILD}
docker-push-all: ${DOCKER_PUSH}
docker-buildx-all-push: ${DOCKER_BUILDX_PUSH}

gen-sdk:
	./hack/sdk/gen-sdk.sh

gen-openapi:
	./hack/gen-openapi.sh

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	export GO111MODULE=on ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
