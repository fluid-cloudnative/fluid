
# Image URL to use all building/pushing image targets
IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/runtime-controller
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

CSI_IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/fluid-csi

LOADER_IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/fluid-dataloader

INIT_USERS_IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/init-users

LOCAL_FLAGS ?= -gcflags=-l
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CURRENT_DIR=$(shell pwd)
VERSION=0.4.0
BUILD_DATE=$(shell date -u +'%Y-%m-%d-%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
GIT_SHA=$(shell git rev-parse --short HEAD || echo "HEAD")
GIT_VERSION=${VERSION}-${GIT_SHA}
PACKAGE=github.com/fluid-cloudnative/fluid

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE} \
  -extldflags "-static"

all: manager

# Run tests
test: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go list ./... | grep -v controller | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# used in CI and simply ignore controller tests which need k8s now.
# maybe incompatible if more end to end tests are added.
unit-test: generate fmt vet
	GO111MODULE=off go list ./... | grep -v controller | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# Build manager binary
manager: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -o bin/manager cmd/controller/main.go

# Build CSI binary
csi: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -o bin/csi -ldflags '${LDFLAGS}' cmd/csi/main.go

# Build CSI binary in dockerfile
csi-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -o /go/bin/fluid-csi -ldflags '${LDFLAGS}' cmd/csi/main.go

# Debug against the configured Kubernetes cluster in ~/.kube/config, add debug
debug: generate fmt vet manifests
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  dlv debug --headless --listen ":12345" --log --api-version=2 cmd/controller/main.go

# Debug against the configured Kubernetes cluster in ~/.kube/config, add debug
debug-csi: generate fmt vet manifests
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  dlv debug --headless --listen ":12346" --log --api-version=2 cmd/csi/main.go -- --nodeid=cn-hongkong.172.31.136.194 --endpoint=unix://var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go run cmd/controller/main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	GO111MODULE=off $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	GO111MODULE=off go fmt ./...

# Run go vet against code
vet:
	GO111MODULE=off go list ./... | grep -v "vendor" | xargs go vet

# Generate code
generate: controller-gen
	GO111MODULE=off $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: generate fmt vet
	docker build --no-cache . -t ${IMG}:${GIT_VERSION}

docker-build-csi: generate fmt vet
	docker build --no-cache . -f Dockerfile.csi -t ${CSI_IMG}:${GIT_VERSION}

docker-build-loader:
	docker build --no-cache charts/fluid-dataloader/docker/loader -t ${LOADER_IMG}

docker-build-init-users:
	docker build --no-cache charts/alluxio/docker/init-users -t ${INIT_USERS_IMG}:${GIT_VERSION}

# Push the docker image
docker-push: docker-build
	docker push ${IMG}:${GIT_VERSION}

docker-push-csi: docker-build-csi
	docker push ${CSI_IMG}:${GIT_VERSION}

docker-push-loader: docker-build-loader
	docker push ${LOADER_IMG}

docker-push-init-users: docker-build-init-users
	docker push ${INIT_USERS_IMG}:${GIT_VERSION}

docker-push-all: docker-push-init-users docker-push docker-push-csi
docker-build-all: docker-build-init-users docker-build docker-build-csi

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
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
