#!/usr/bin/env bash

# Stops the server.

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

if [[ "$test_instance" == "true" ]]; then
    root="testroot"

    if [[ -e "$root" ]]; then
        rm -rf "$root"
    fi
fi

docker compose down -v