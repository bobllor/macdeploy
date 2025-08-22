#!/usr/bin/env bash

# Build the Docker images.
# Must run with sudo OR the current user has the "docker" group assigned.

fs_target="fsserver"
go_target="gopipe"
cron_target="cronner"

args=($fs_target)

while (( $# > 0 )); do
    case "$1" in
        --action )
            shift
            if [[ ! -e "./.github/workflows" ]]; then
                echo "no .github/workflows directory found"
            elif [[ -z $(find ./.github/workflows/ -name "*.yml") ]]; then
                echo "no actions YAML found" 
            else
                args+=($go_target)
            fi
            ;;
        * )
            echo "invalid option"
            exit 1
            ;;
    esac
    shift
done

# i want to keep the order of how it is in the dockerfile.
args+=($cron_target)

for var in "${args[@]}"; do
    docker build . --target $var -t deploy:$var
done