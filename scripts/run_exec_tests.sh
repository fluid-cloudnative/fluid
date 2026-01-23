#!/bin/bash
#
# Quick script to run just the exec.go tests with race detection
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/.."

echo "Installing dependencies..."
go mod download

echo ""
echo "Running exec.go tests with race detection..."
echo "============================================"

go test -race -v -count=1 ./pkg/utils/kubeclient/... \
    -run "TestExecCommandInContainerWithTimeout|TestExecResult|TestInitClient"

echo ""
echo "============================================"
echo "All tests passed!"
