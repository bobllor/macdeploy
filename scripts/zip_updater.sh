#!/usr/bin/env bash

# Accesses the endpoint that triggers the ZIP update API.
# This should only be used by the container.

# probably not needed but ill just keep it here just in case.
cd /macos-deployment/server

token=$(cat .token)
host="python-fs:5000" # this must match the name of the compose.yml name.

curl -H "x-zip-token: $token" https://$host/api/zip/update --insecure \
    &>> /tmp/cronner.log