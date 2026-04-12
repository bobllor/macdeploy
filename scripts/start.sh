#!/usr/bin/env bash

# Starts the server.
#
# If the flag '-t' is used, the test instance is created with test folders.
# This is not intended to be used with normal development, but used for the Docker
# container.

help(){
    value="Usage: $0 [-h] [-t]"

    echo "$value"
    exit 1
}

test_instance=false
while getopts "th" arg; do
    case "${arg}" in
        t) test_instance=true ;;
        h) help ;;
        *) ;;
    esac
done

if [[ "$test_instance" == "false" ]]; then
    docker compose up -d
else
    root="testroot"

    mkdir -p "$root"

    key_path="$root/keys"

    files=("$root/data" "$key_path" "$root/logs" "$root/zip-build")

    for file in "${files[@]}"; do
        mkdir -p "$file"

        if [[ "$file" =~ "keys" ]]; then
            serial="$file/SERIALTAG1"
            mkdir -p "$serial"
            
            touch "$serial/ABC1-23DF-GH45-J6KL-7MN8-OPQ9"
        fi
    done

    # separate serial entry due to github actions and permission issues
    # used for testing adding key entries to the folder.
    mkdir "$key_path/SERIALTAG3"

    # this is only related to testing due to permission errors due to github actions
    chmod -R 777 "$root"

    docker compose -f compose.yml -f dockerfiles/compose-test.override.yml up -d
fi