#!/usr/bin/env bash

# Build the Docker images.
# Must run with sudo OR the current user has the "docker" group assigned.

fs_target="fsserver"
go_target="gopipe" # TODO: not using yet
cron_target="cronner"

args=($fs_target $cron_target)

for var in "${args[@]}"; do
    docker build . --target $var -t deploy:$var
done