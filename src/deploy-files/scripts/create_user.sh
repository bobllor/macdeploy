#!/bin/bash

# creates the user on the macOS.
# by default it creates a non-admin user.
# args:
#   - 

# $0/$user_name if used with the FormatName function in my Go script, will be in title case already.
# used for the -fullName flag, this is the display name of the user.
user_name=$0
folder_name=$1
password=$2
isAdmin=$3

if [[ $isAdmin == "false" ]]; then
    sudo sysadminctl -addUser "$folder_name" \
        -fullName "$user_name" -password "$password"
else
    sudo sysadminctl -addUser "$folder_name" \
        -fullName "$full_name" -password "$password" -admin
fi