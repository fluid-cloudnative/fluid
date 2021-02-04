
# Image URL to use all building/pushing image targets
IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/runtime-controller
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# The Image URL to use in docker build and push
DATASET_CONTROLLER_IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/dataset-controller
ALLUXIORUNTIME_CONTROLLER_IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/alluxioruntime-controller
JINDORUNTIME_CONTROLLER_IMG ?= registry.cn-hangzhou.aliyuncs.com/fluid/jindoruntime-controller
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
VERSION=v0.5.0
BUILD_DATE=$(shell date -u +'%Y-%m-%d_%H:%M:%S')
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

all: build

# Run tests
test: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go list ./... | grep -v controller | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# used in CI and simply ignore controller tests which need k8s now.
# maybe incompatible if more end to end tests are added.
unit-test: generate fmt vet
	GO111MODULE=off go list ./... | grep -v controller | xargs go test ${CI_TEST_FLAGS} ${LOCAL_FLAGS}

# Build binary

build: dataset-controller-build alluxioruntime-controller-build jindoruntime-controller-build csi-build

csi-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -o bin/fluid-csi -ldflags '${LDFLAGS}' cmd/csi/main.go

dataset-controller-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -gcflags="-N -l" -a -o bin/dataset-controller -ldflags '${LDFLAGS}' cmd/dataset/main.go

alluxioruntime-controller-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -gcflags="-N -l" -a -o bin/alluxioruntime-controller -ldflags '${LDFLAGS}' cmd/alluxio/main.go

jindoruntime-controller-build: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off  go build -gcflags="-N -l" -a -o bin/jindoruntime-controller -ldflags '${LDFLAGS}' cmd/jindo/main.go

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

# Run go fmt against codecsi-node-driver-registrar
fmt:
	GO111MODULE=off go fmt ./...

# Run go vet against code
vet:
	GO111MODULE=off go list ./... | grep -v "vendor" | xargs go vet

# Generate code
generate: controller-gen
	GO111MODULE=off $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Update fluid helm chart
update-crd: manifests
	cp config/crd/bases/* charts/fluid/fluid/crds

update-api-doc:
	bash tools/api-doc-gen/generate_api_doc.sh && mv tools/api-doc-gen/api_doc.md docs/zh/dev/api_doc.md && cp docs/zh/dev/api_doc.md docs/en/dev/api_doc.md

# Build the docker image
docker-build-dataset-controller: generate fmt vet
	docker build --no-cache . -f Dockerfile.dataset -t ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-alluxioruntime-controller: generate fmt vet
	docker build --no-cache . -f Dockerfile.alluxioruntime -t ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-jindoruntime-controller: generate fmt vet
	docker build --no-cache . -f Dockerfile.jindoruntime -t ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-build-csi: generate fmt vet
	docker build --no-cache . -f Dockerfile.csi -t ${CSI_IMG}:${GIT_VERSION}

docker-build-loader:
	docker build --no-cache charts/fluid-dataloader/docker/loader -t ${LOADER_IMG}

docker-build-init-users:
	docker build --no-cache charts/alluxio/docker/init-users -t ${INIT_USERS_IMG}:${GIT_VERSION}

# Push the docker image
docker-push-dataset-controller: docker-build-dataset-controller
	docker push ${DATASET_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-alluxioruntime-controller: docker-build-alluxioruntime-controller
	docker push ${ALLUXIORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-jindoruntime-controller: docker-build-jindoruntime-controller
	docker push ${JINDORUNTIME_CONTROLLER_IMG}:${GIT_VERSION}

docker-push-csi: docker-build-csi
	docker push ${CSI_IMG}:${GIT_VERSION}

docker-push-loader: docker-build-loader
	docker push ${LOADER_IMG}

docker-push-init-users: docker-build-init-users
	docker push ${INIT_USERS_IMG}:${GIT_VERSION}

docker-build-all: docker-build-dataset-controller docker-build-alluxioruntime-controller docker-build-jindoruntime-controller docker-build-csi docker-build-init-users
docker-push-all: docker-push-dataset-controller docker-push-alluxioruntime-controller docker-push-jindoruntime-controller docker-push-csi docker-push-init-users

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
