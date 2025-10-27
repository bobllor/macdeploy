#!/usr/bin/env bash

# All-in-one build script.
# Can be used for initializing the project and rebuilding containers.

set -e

source "scripts/utils/utils.sh"

zip=false
include_x86=false
while getopts "zx" o; do
    case "${o}" in
        z)
            zip=true
            ;;
        x)
            include_x86=true
            ;;
        *) ;;
    esac
done

# expected to be ran in the root directory
filename="scripts/filename.sh"

keys_var="KEYS_NAME"
keys_dir=$(filename "$keys_var")
varcheck "$keys_dir" "$keys_var" || exit 1

logs_var="LOGS_NAME"
logs_dir=$(filename "$logs_var")
varcheck "$logs_dir" "$logs_var" || exit 1

dist_var="DIST_DIR_NAME"
dist_dir=$(filename "$dist_var")
varcheck "$dist_dir" "$dist_var" || exit 1

zip_dir_var="ZIP_DIR_NAME"
zip_dir=$(filename "$zip_dir_var")
varcheck "$zip_dir" "$zip_dir_var" || exit 1

mkdir -p $keys_dir
mkdir -p $logs_dir
mkdir -p $dist_dir
mkdir -p $zip_dir

if [[ $zip == true ]]; then
    config=$(ls -t | grep -Ei "^config\.(yaml|yml)$" | head -1)
    if [[ -z "$config" ]]; then
        echo "No YAML config found"
    else
        zip_script="go_zip.sh"

        if [[ $include_x86 == false ]]; then
            bash "scripts/$zip_script"
        else
            bash "scripts/$zip_script" -x
        fi
    fi
fi

docker compose down -v
docker compose build && docker compose create