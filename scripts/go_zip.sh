#!/usr/bin/env bash

# Creates the binary and updates the ZIP package in the root directory.
# Used if a new binary is required for an update, for example a change in the YAML config.

# getting the latest changed yaml config in the server.
configs=$(ls -t | grep -Ei "config\.(yaml|yml)$")
config=""

while read line; do
    config=$line
    break
done <<< $configs

if [[ -z "$config" ]]; then
    echo "No YAML config found"
    exit 1
fi

dist_dir="dist"

if [[ ! -d "$dist_dir" ]]; then
    mkdir $dist_dir
fi

dest_config="config.yml"

# copies the YAML config into src for embedding
cp "$config" "./src/config/$dest_config"

binary_name="macdeploy"
env GOOS=darwin GOARCH=arm64 go build -C ./src -o "../dist/$binary_name"
echo "binary output: dist/$binary_name"

amd_binary="x86_64-macdeploy"
env GOOS=darwin GOARCH=amd64 go build -C ./src -o "../dist/$amd_binary"
echo "binary output: dist/$amd_binary"

zip_name="deploy.zip"

echo "generating zip file"
zip -ru "$zip_name" "$dist_dir"