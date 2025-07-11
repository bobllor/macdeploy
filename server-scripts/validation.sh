#!/bin/bash

# validation for server files
deployment=/Users/$(whoami)/mac-deployment

if [[ ! -e $deployment/logs ]]; then
	mkdir $deployment/logs
fi

# if pkg-files is missing then prevent the script from executing
if [[ ! -e $deployment/pkg-files ]]; then
  echo "pkg-files directory is missing, exiting script"
  exit 1
fi

echo 0
