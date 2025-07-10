#!/bin/bash

# validation for server files
deployment=/Users/$(whoami)/mac-deployment

if [[ ! -e $deployment/logs ]]; then
	mkdir $deployment/logs
fi
