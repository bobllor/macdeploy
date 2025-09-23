#!/bin/bash

# Creates the user, by default it creates a non-admin user.
# Args:
#   - 0: The user's display name, or Full Name.
#   - 1: The account name used internally.
#   - 2: The user's password.
#   - 3: String boolean used to grant admin to the user.

full_name=$0
account_name=$1
password=$2
isAdmin=$3

if [[ $isAdmin == "false" ]]; then
    sudo sysadminctl -addUser "$account_name" \
        -fullName "$full_name" -password "$password"
else
    sudo sysadminctl -addUser "$account_name" \
        -fullName "$full_name" -password "$password" -admin
fi