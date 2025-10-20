#!/usr/bin/env bash

# Creates the binary and updates the ZIP package in the root directory.
# Used if a new binary is required for an update, for example a change in the YAML config.
# This must be ran in the root directory.

set -e

source "scripts/utils/utils.sh"

# getting the latest changed yaml config in the server.
config=$(ls -t | grep -Ei "^config\.(yaml|yml)$" | head -1)
if [[ -z "$config" ]]; then
    echo "No YAML config found"
    exit 1
fi

dist_var="DIST_DIR_NAME"
dist_dir=$(filename "$dist_var")
varcheck "$dist_dir" "$dist_var" || exit 1

mkdir -p $dist_dir

# used for embedding in go.
dest_config="config.yml"

# copies the YAML config into src for embedding
cp "$config" "./src/config/$dest_config"

zip_var="ZIP_NAME"
zip_name=$(filename "$zip_var")
varcheck "$zip_name" "$zip_var" || exit 1

zip_dir_var="ZIP_DIR_NAME"
zip_dir=$(filename "$zip_dir_var")
varcheck "$zip_dir" "$zip_dir_var" || exit 1

mkdir -p $zip_dir

binary_name="macdeploy"
env GOOS=darwin GOARCH=arm64 go build -C ./src -o "../dist/$binary_name"
printf "Binary output: dist/$binary_name\n"

amd_binary="x86_64-macdeploy"
env GOOS=darwin GOARCH=amd64 go build -C ./src -o "../dist/$amd_binary"
printf "Binary output: dist/$amd_binary\n"

printf "Generating zip file\n"
zip -ru "$zip_dir/$zip_name" "$dist_dir"