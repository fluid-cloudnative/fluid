CURRENT_DIR=$(shell pwd)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	@hack/generate_client.sh

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
ifeq ("$(shell $(CONTROLLER_GEN) --version 2> /dev/null)", "Version: v0.7.0")
else
	rm -rf $(CONTROLLER_GEN)
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)
endif

OPENAPI_GEN = $(shell pwd)/bin/openapi-gen
module=$(shell go list -f '{{.Module}}' k8s.io/kube-openapi/cmd/openapi-gen | awk '{print $$1}')
module_version=$(shell go list -m $(module) | awk '{print $$NF}' | head -1)
openapi-gen: ## Download openapi-gen locally if necessary.
ifeq ("$(shell command -v $(OPENAPI_GEN) 2> /dev/null)", "")
	$(call go-get-tool,$(OPENAPI_GEN),k8s.io/kube-openapi/cmd/openapi-gen@$(module_version))
else
	@echo "openapi-gen is already installed."
endif

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: gen-schema-only
gen-schema-only:
	go run cmd/gen-schema/main.go

.PHONY: gen-openapi-schema
gen-openapi-schema: gen-all-openapi gen-kruise-openapi gen-rollouts-openapi
	go run cmd/gen-schema/main.go

.PHONY: gen-all-openapi
gen-all-openapi: openapi-gen
	$(OPENAPI_GEN) \
	  	--go-header-file ./hack/boilerplate.go.txt \
		--input-dirs github.com/openkruise/kruise-api/apps/v1alpha1,github.com/openkruise/kruise-api/apps/pub,github.com/openkruise/kruise-api/apps/v1beta1,github.com/openkruise/kruise-api/policy/v1alpha1,github.com/openkruise/kruise-api/rollouts/v1alpha1,k8s.io/api/admission/v1,k8s.io/api/admissionregistration/v1,k8s.io/api/admissionregistration/v1beta1,k8s.io/api/authentication/v1,k8s.io/api/apps/v1,k8s.io/api/apps/v1beta1,k8s.io/api/apps/v1beta2,k8s.io/api/autoscaling/v1,k8s.io/api/batch/v1,k8s.io/api/batch/v1beta1,k8s.io/api/certificates/v1beta1,k8s.io/api/certificates/v1,k8s.io/api/core/v1,k8s.io/api/extensions/v1beta1,k8s.io/api/networking/v1,k8s.io/api/networking/v1beta1,k8s.io/api/policy/v1,k8s.io/api/policy/v1beta1,k8s.io/api/rbac/v1,k8s.io/api/rbac/v1alpha1,k8s.io/api/storage/v1,k8s.io/api/storage/v1alpha1,k8s.io/api/storage/v1beta1 \
		--output-package ./pkg/apis \
  		--report-filename ./pkg/apis/violation_exceptions.list \
  		-o $(CURRENT_DIR)

.PHONY: gen-kruise-openapi
gen-kruise-openapi: openapi-gen
	$(OPENAPI_GEN) \
	  	--go-header-file hack/boilerplate.go.txt \
		--input-dirs github.com/openkruise/kruise-api/apps/v1alpha1,github.com/openkruise/kruise-api/apps/pub,github.com/openkruise/kruise-api/apps/v1beta1,github.com/openkruise/kruise-api/policy/v1alpha1 \
		--output-package pkg/kruise/ \
  		--report-filename pkg/kruise/violation_exceptions.list \
  		-o $(CURRENT_DIR)

.PHONY: gen-rollouts-openapi
gen-rollouts-openapi: openapi-gen
	$(OPENAPI_GEN) \
	  	--go-header-file hack/boilerplate.go.txt \
		--input-dirs github.com/openkruise/kruise-api/rollouts/v1alpha1 \
		--output-package pkg/rollouts/ \
  		--report-filename pkg/rollouts/violation_exceptions.list \
  		-o $(CURRENT_DIR)
