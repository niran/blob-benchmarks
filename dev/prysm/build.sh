#!/bin/bash

# Check if path argument is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <path-to-prysm-repo>"
    exit 1
fi

# Convert relative path to absolute path
PRYSM_REPO_PATH=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BIN_DIR="$SCRIPT_DIR/bin"

# Check if prysm repo exists at provided path
if [ ! -d "$PRYSM_REPO_PATH" ]; then
    echo "Error: Prysm repository not found at $PRYSM_REPO_PATH"
    exit 1
fi

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

# Build the binaries using Docker
echo "Building Prysm binaries..."
docker run -v "$PRYSM_REPO_PATH":/workspace -v "$BIN_DIR":/output -w /workspace --rm -it golang:1.23-bookworm /bin/bash -c '
    go build -v -o /output/beacon-chain ./cmd/beacon-chain && \
    go build -v -o /output/validator ./cmd/validator
'

if [ $? -eq 0 ]; then
    echo "Successfully built binaries in $BIN_DIR"
    
    # Build Docker images
    echo "Building Docker images..."
    cd "$SCRIPT_DIR"
    docker build -t prysm-beacon-chain-dev -f Dockerfile.beacon-chain .
    docker build -t prysm-validator-dev -f Dockerfile.validator .
    
    echo "Build process completed successfully!"
else
    echo "Error: Failed to build Prysm binaries"
    exit 1
fi 
