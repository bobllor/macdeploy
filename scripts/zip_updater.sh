#!/usr/bin/env bash

# Accesses the endpoint that triggers the ZIP update API.
# This should only be used by the container.

cd /macos-deployment/src/server

token=$(cat .token)
host="python-fs:5000" # this must match the name of the compose.yml name.

curl -H "x-zip-token: $token" https://$host/api/zip/update --insecure \
    &>> /tmp/cronner.log