#!/usr/bin/env bash

# Copyright 2019 The Kubeflow Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SWAGGER_CODEGEN_JAR="hack/sdk/swagger-codegen-cli.jar"
SWAGGER_CODEGEN_CONF="hack/sdk/swagger_config.json"
SWAGGER_CODEGEN_FILE="api/v1alpha1/swagger.json"
SDK_OUTPUT_PATH="sdk/python"
PYTHON_SDK_OUTPUT_PATH="sdk/python"
JAVA_SDK_OUTPUT_PATH="sdk/java"

if [ -z "${GOPATH:-}" ]; then
    export GOPATH=$(go env GOPATH)
fi

# Grab kube-openapi version from go.mod
OPENAPI_VERSION=$(grep 'k8s.io/kube-openapi' go.mod | awk '{print $2}' | head -1)

echo "Installing openapi-gen"
go install k8s.io/kube-openapi/cmd/openapi-gen@${OPENAPI_VERSION}

echo "Generating OpenAPI specification ..."
${GOPATH}/bin/openapi-gen --input-dirs github.com/fluid-cloudnative/fluid/api/v1alpha1 --output-package github.com/fluid-cloudnative/fluid/api/v1alpha1 --go-header-file hack/boilerplate.go.txt

echo "Generating swagger file ..."
go run hack/sdk/main.go 0.1 > ${SWAGGER_CODEGEN_FILE}

echo "Generating python SDK for Fluid ..."
java -jar ${SWAGGER_CODEGEN_JAR} generate -i ${SWAGGER_CODEGEN_FILE} -l python -o ${PYTHON_SDK_OUTPUT_PATH} -c ${SWAGGER_CODEGEN_CONF} --model-package com.github.fluid-cloudnative.fluid

echo "Fluid Python SDK is generated successfully to folder ${PYTHON_SDK_OUTPUT_PATH}/."

echo "Generating java SDK for Fluid ..."
java -jar ${SWAGGER_CODEGEN_JAR} generate -i ${SWAGGER_CODEGEN_FILE} -l java -o ${JAVA_SDK_OUTPUT_PATH} -c ${SWAGGER_CODEGEN_CONF} --model-package com.github.fluid-cloudnative.fluid

echo "Fluid Java SDK is generated successfully to folder ${JAVA_SDK_OUTPUT_PATH}/."


