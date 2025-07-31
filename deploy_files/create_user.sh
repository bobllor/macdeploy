#!/bin/bash

# creates the user on the macOS.
# by default it creates a non-admin user.
# args:
#   - 

user_name=$1
full_name=$2
password=$3
isAdmin=$4

if [[ isAdmin == "no" ]]; then
    sudo sysadminctl -addUser "$user_name" \
        -fullName "$full_name" -password "$password"
else
    sudo sysadminctl -addUser "$user_name" \
        -fullName "$full_name" -password "$password" -admin
fi
