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

# Combine base YAMLs and provided arguments
files=()

# Add kurtosis/base.yaml if it exists
BASE_YAML="kurtosis/base.yaml"
[ -f "$BASE_YAML" ] && files+=("$BASE_YAML")

# Look for 000-base.yaml in the directory of the first argument that has one
for arg in "$@"; do
    dir_base="$(dirname "$arg")/000-base.yaml"
    if [ -f "$dir_base" ] && ! echo "$*" | grep -q "$dir_base"; then
        files+=("$dir_base")
        break
    fi
done

# Add all provided arguments
files+=("$@")

# Merge all YAML files using yq
for file in "${files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "Error: File $file does not exist"
        exit 1
    fi
    echo "Merging configuration from $file"
    yq '. *= load("'"$file"'")' "$MERGED_ARGS_FILE" > "${MERGED_ARGS_FILE}.tmp" && mv "${MERGED_ARGS_FILE}.tmp" "$MERGED_ARGS_FILE"
done

# Run kurtosis with the merged args file
kurtosis run github.com/ethpandaops/ethereum-package --args-file "$MERGED_ARGS_FILE" 
