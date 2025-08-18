#!/usr/bin/env bash

# Creates the ZIP file. If there is an existing ZIP file, then update it.
# This should only be ran once upon project load, but can also be used to remake it
# manually outside of the Docker containers if needed.

deploy_dir="deploy-zip"
pkg_dir="pkg-files"

if [[ ! -d "$deploy_dir" ]]; then
    mkdir $deploy_dir
fi
if [[ ! -d "$pkg_dir" ]]; then
    mkdir $pkg_dir
fi

zip_name="deploy.zip"
go_bin="deploy.bin"
config="config.yaml"

zip -ru "$deploy_dir"/"$zip_name" $pkg_dir $go_bin $config