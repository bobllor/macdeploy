#!/usr/bin/env bash

# Creates the binary and updates the ZIP package.
# Used if a new binary is required for an update, for example a change in the YAML config.

dist_dir="dist"

if [[ ! -d "$dist_dir" ]]; then
    mkdir $dist_dir
fi

config_name="config.yml"
# used for the destination copy, handles .yaml and .yml
dest_config=$config_name

if [[ ! -e "$config_name" ]]; then
    alt_config_name="config.yaml"

    if [[ ! -e "$alt_config_name" ]]; then
        echo "cannot find YAML config file"
        exit 1
    fi

    config_name=$alt_config_name
fi

cp $config_name ./src/config/$dest_config

binary_name="deploy-arm.bin"
env GOOS=darwin GOARCH=arm64 go build -C ./src -o "../dist/$binary_name"

# need another binary for intel based macs
amd_binary="deploy-x86_64.bin"
env GOOS=darwin GOARCH=amd64 go build -C ./src -o "../dist/$amd_binary"

pkg_name="pkg-files"
pkg_dir="$dist_dir/$pkg_name"

if [[ ! -d "$pkg_dir" ]]; then
    mkdir $pkg_dir
fi

cd $dist_dir
zip_name="deploy.zip"

zip -ru "$zip_name" "$pkg_name" "$binary_name" "$amd_binary"