#!/usr/bin/env bash

# Accesses the endpoint that triggers the ZIP update API.
# This should only be used by the container.

cd /macdeploy/src/server

log="/tmp/cronner.log"

log_word_count=$(cat $log | wc -l)
if [[ $log_word_count -gt 100 ]]; then
    log_file_count=$(ls /tmp | grep -E "cronner\.*" | wc -l)
    log="/tmp/cronner-$log_file_count.log"
fi

token=$(cat .token)
host="python-fs:5000" # this must match the name of the compose.yml name.

curl -H "x-zip-token: $token" https://$host/api/zip/update --insecure \
    &>> $log