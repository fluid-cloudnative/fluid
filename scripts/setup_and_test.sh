#!/bin/bash
#
# Setup script to install dependencies and run tests for Fluid
# This script installs Go (if not present), downloads dependencies,
# and runs the tests with race detection.
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN} Fluid Test Setup Script${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if Go is installed
check_go() {
    echo -e "\n${YELLOW}Checking Go installation...${NC}"
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version)
        echo -e "${GREEN}Go is installed: ${GO_VERSION}${NC}"
        return 0
    else
        echo -e "${RED}Go is not installed.${NC}"
        return 1
    fi
}

# Install Go
install_go() {
    echo -e "\n${YELLOW}Installing Go 1.22.x...${NC}"
    
    # Download Go
    GO_VERSION="1.22.5"
    GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
    
    echo "Downloading Go ${GO_VERSION}..."
    wget -q "https://go.dev/dl/${GO_TAR}" -O "/tmp/${GO_TAR}"
    
    echo "Extracting Go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/${GO_TAR}"
    rm "/tmp/${GO_TAR}"
    
    # Add to PATH if not already there
    if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    fi
    
    # Export for current session
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    
    echo -e "${GREEN}Go ${GO_VERSION} installed successfully!${NC}"
}

# Download Go module dependencies
download_dependencies() {
    echo -e "\n${YELLOW}Downloading Go module dependencies...${NC}"
    
    cd "$(dirname "$0")/.."
    
    echo "Running go mod download..."
    go mod download
    
    echo "Running go mod tidy..."
    go mod tidy
    
    echo -e "${GREEN}Dependencies downloaded successfully!${NC}"
}

# Run tests
run_tests() {
    echo -e "\n${YELLOW}Running tests with race detection...${NC}"
    
    cd "$(dirname "$0")/.."
    
    echo -e "\n${YELLOW}Testing pkg/utils/kubeclient/exec.go...${NC}"
    go test -race -v ./pkg/utils/kubeclient/... -run "ExecCommandInContainerWithTimeout|ExecResult"
    
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN} All tests passed!${NC}"
    echo -e "${GREEN}========================================${NC}"
}

# Run all kubeclient tests
run_all_kubeclient_tests() {
    echo -e "\n${YELLOW}Running all kubeclient tests...${NC}"
    
    cd "$(dirname "$0")/.."
    
    go test -race -v ./pkg/utils/kubeclient/...
}

# Main script
main() {
    # Change to fluid directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "${SCRIPT_DIR}/.."
    
    echo "Working directory: $(pwd)"
    
    # Check and install Go if needed
    if ! check_go; then
        install_go
    fi
    
    # Verify Go is working
    go version
    
    # Download dependencies
    download_dependencies
    
    # Run tests
    run_tests
    
    # Optionally run all kubeclient tests
    read -p "Do you want to run all kubeclient tests? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        run_all_kubeclient_tests
    fi
}

main "$@"
