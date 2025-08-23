#!/usr/bin/env bash

# Creates the binary and updates the ZIP package.
# Used if a new binary is required for an update, for example a change in the YAML config.

dist_dir="dist"

if [[ ! -d "$dist_dir" ]]; then
    mkdir $dist_dir
fi

binary_name="deploy.bin"

cp config.yml ./src/config 
env GOOS=darwin GOARCH=arm64 go build -C ./src -o "../dist/$binary_name"

pkg_name="pkg-files"
pkg_dir="$dist_dir/$pkg_name"

if [[ ! -d "$pkg_dir" ]]; then
    mkdir $pkg_dir
fi

cd $dist_dir
zip_name="deploy.zip"

zip -ru "$zip_name" $pkg_name $binary_name