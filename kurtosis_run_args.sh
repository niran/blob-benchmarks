#!/bin/bash

# Check if yq is installed
if ! command -v yq &> /dev/null; then
    echo "Error: yq is required but not installed."
    echo "Install it with: brew install yq"
    exit 1
fi

# Initialize arrays for yaml files and kurtosis flags
yaml_files=()
kurtosis_flags=()

# Parse arguments - collect consecutive non-flag arguments at the start as YAML files
while [[ $# -gt 0 ]]; do
    # Break if we hit a flag (starts with -)
    if [[ "$1" == -* ]]; then
        break
    fi
    yaml_files+=("$1")
    shift
done

# Remaining arguments are flags
kurtosis_flags=("$@")

# Check if at least one YAML file is provided
if [ ${#yaml_files[@]} -eq 0 ]; then
    echo "Usage: $0 yaml_file1 [yaml_file2 ...] [flags]"
    echo "Note: YAML files must come first, followed by any flags for 'kurtosis run'"
    exit 1
fi

# Validate that all yaml_files actually end in .yaml
for file in "${yaml_files[@]}"; do
    if [[ ! "$file" == *.yaml ]]; then
        echo "Error: '$file' is not a YAML file"
        echo "Usage: $0 yaml_file1 [yaml_file2 ...] [flags]"
        echo "Note: YAML files must come first, followed by any flags for 'kurtosis run'"
        exit 1
    fi
done

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

# Look for 000-base.yaml in the directory of the first yaml file that has one
for yaml_file in "${yaml_files[@]}"; do
    dir_base="$(dirname "$yaml_file")/000-base.yaml"
    if [ -f "$dir_base" ] && ! printf "%s\n" "${yaml_files[@]}" | grep -q "^$dir_base$"; then
        files+=("$dir_base")
        break
    fi
done

# Add all provided YAML files
files+=("${yaml_files[@]}")

# Merge all YAML files using yq
for file in "${files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "Error: File $file does not exist"
        exit 1
    fi
    echo "Merging configuration from $file"
    yq '. *= load("'"$file"'")' "$MERGED_ARGS_FILE" > "${MERGED_ARGS_FILE}.tmp" && mv "${MERGED_ARGS_FILE}.tmp" "$MERGED_ARGS_FILE"
done

# Run kurtosis with the merged args file and any additional flags
kurtosis run github.com/ethpandaops/ethereum-package --args-file "$MERGED_ARGS_FILE" "${kurtosis_flags[@]}" 
