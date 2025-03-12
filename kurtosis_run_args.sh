#!/bin/bash

# Check if yq is installed
if ! command -v yq &> /dev/null; then
    echo "Error: yq is required but not installed."
    echo "Install it with: brew install yq"
    exit 1
fi

# Check if at least one argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 yaml_file1 [yaml_file2 ...]"
    exit 1
fi

# Create a temporary file for the merged YAML
MERGED_ARGS_FILE=$(mktemp)

# Ensure temporary file is deleted on script exit
trap "rm -f $MERGED_ARGS_FILE" EXIT

# Initialize with empty YAML
echo "{}" > "$MERGED_ARGS_FILE"

# Merge all YAML files using yq
for file in "$@"; do
    if [ ! -f "$file" ]; then
        echo "Error: File $file does not exist"
        exit 1
    fi
    yq '. *= load("'"$file"'")' "$MERGED_ARGS_FILE" > "${MERGED_ARGS_FILE}.tmp" && mv "${MERGED_ARGS_FILE}.tmp" "$MERGED_ARGS_FILE"
done

# Run kurtosis with the merged args file
kurtosis run github.com/ethpandaops/ethereum-package --args-file "$MERGED_ARGS_FILE" 
