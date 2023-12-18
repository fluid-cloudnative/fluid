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
SWAGGER_CODEGEN_JAR_URL="https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/4.3.1/openapi-generator-cli-4.3.1.jar"
SWAGGER_CODEGEN_CONF="hack/sdk/swagger_config.json"
SWAGGER_CODEGEN_FILE="api/v1alpha1/swagger.json"
PYTHON_SDK_OUTPUT_PATH="sdk/python"
JAVA_SDK_OUTPUT_PATH="sdk/java"
POST_GEN_SDK_SCRIPT="hack/sdk/post-gen.py"

if [ -z "${GOPATH:-}" ]; then
    export GOPATH=$(go env GOPATH)
fi

# Grab kube-openapi version from go.mod
OPENAPI_VERSION=$(grep 'k8s.io/kube-openapi' go.mod | awk '{print $2}' | head -1)

echo "Installing openapi-gen"
go install k8s.io/kube-openapi/cmd/openapi-gen@${OPENAPI_VERSION}

echo "Generating OpenAPI specification ..."
${GOPATH}/bin/openapi-gen --input-dirs github.com/fluid-cloudnative/fluid/api/v1alpha1 --output-package github.com/fluid-cloudnative/fluid/api/v1alpha1 --go-header-file hack/boilerplate.go.txt

echo "Downloading codegen jar ..."
if [ -f "${SWAGGER_CODEGEN_JAR}" ]; then
    echo "Using existing ${SWAGGER_CODEGEN_JAR}"
else
    if ! command -v curl >/dev/null 2>&1; then
        echo "Error: curl command not found." >&2
        exit 1
    fi
    for i in {1..3}; do
        if wget -O "${SWAGGER_CODEGEN_JAR}" "${SWAGGER_CODEGEN_JAR_URL}"; then
            break
        elif [ "$i" -eq 3 ]; then
            echo "Failed to download ${SWAGGER_CODEGEN_JAR} after 3 attempts." >&2
            exit 1
        else
            echo "Failed to download ${SWAGGER_CODEGEN_JAR}, retrying in 10 seconds..." >&2
            sleep 10
        fi
    done
fi

echo "Generating swagger file ..."
go run hack/sdk/main.go 0.1 > ${SWAGGER_CODEGEN_FILE}

echo "Generating python SDK for Fluid ..."
java -jar ${SWAGGER_CODEGEN_JAR} generate -i ${SWAGGER_CODEGEN_FILE} -g python -o ${PYTHON_SDK_OUTPUT_PATH} -c ${SWAGGER_CODEGEN_CONF} --model-package models
# Revert files that are diverged from the generated files
pushd . && \
    cd ${PYTHON_SDK_OUTPUT_PATH} && \
    git checkout setup.py && \
    git checkout requirements.txt && \
    git checkout README.md && \
    git checkout fluid/__init__.py && \
    git checkout .gitignore && \
    popd

python3 ${POST_GEN_SDK_SCRIPT} --python-sdk-path=${PYTHON_SDK_OUTPUT_PATH}

echo "Fluid Python SDK is generated successfully to folder ${PYTHON_SDK_OUTPUT_PATH}/."

echo "Generating java SDK for Fluid ..."
java -jar ${SWAGGER_CODEGEN_JAR} generate -i ${SWAGGER_CODEGEN_FILE} -g java -o ${JAVA_SDK_OUTPUT_PATH} -c ${SWAGGER_CODEGEN_CONF} --model-package models

echo "Fluid Java SDK is generated successfully to folder ${JAVA_SDK_OUTPUT_PATH}/."


