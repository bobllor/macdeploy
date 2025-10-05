#/usr/bin/env bash

# Create the necessary files and directories for the server.
# This is only required on the first run prior to running docker compose.

mkdir logs
mkdir dist
mkdir keys

bash scripts/go_zip.sh