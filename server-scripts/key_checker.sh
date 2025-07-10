#!/bin/bash

# script used to check for any duplicate keys and remove them from the folder.
# this script is only for ssh, this is not intended to be used on the client.
# logs are echoed in this and must be captured by the client, which will
# be moved into the server logs folder.
# 
# args:
# 	i dont know what to put here 

serial_number=$1
recovery_key=$2

if [[ -z $serial_number ]]; then
	echo "No serial number given, exiting script"
	exit 1
elif [[ -z $recovery_key ]]; then
	echo "No recovery key given, exiting script"
	exit 1 
fi

main_dir=filevault-keys
default_path=~/mac-deployment/$main_dir

# sanity check
if [[ ! -e $default_path ]]; then
	mkdir $default_path
fi

key_dir=$default_path/$serial_number

# remove the folder and create a new one.
if [[ ! -e $key_dir ]]; then
	echo "Created an entry for $1"
	mkdir -p $key_dir/$recovery_key
else
	echo "Found existing entry for $1 with key $(ls $key_dir), removing and creating new entry"
	rm -rf $key_dir

	mkdir -p $key_dir/$recovery_key
fi
