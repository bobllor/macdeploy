#!/bin/bash

# creates the user on the macOS.
# by default it creates a non-admin user.
# args:
#   - 

# $1/$user_name if used with the FormatName function in my Go script, will be in title case already.
# used for the -fullName flag, this is the display name of the user.
user_name=$1
password=$2
isAdmin=$3

# used for the -addUser flag, this is the directory of the user.
new_user_name=$(awk '{ print tolower($0) }' <<< $1)

if [[ $isAdmin == "false" ]]; then
    sudo sysadminctl -addUser "$new_user_name" \
        -fullName "$user_name" -password "$password"
else
    sudo sysadminctl -addUser "$user_name" \
        -fullName "$full_name" -password "$password" -admin
fi