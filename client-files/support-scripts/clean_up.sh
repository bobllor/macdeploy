#!/bin/bash

# cleans up the installed and created files on the client device

source ./globals.sh

# in case this is ran on the server
# the function init in the main script has this as well.
if [[ $(whoami) == $user ]]; then
  exit 1
fi

rm ~/*.log

# potentially dangerous!
# IMPORTANT: DO NOT REMOVE THE CONDITIONAL CHECK.
if [[ -e ~/$script_dir ]]; then
  echo "Removing installed scripts"
  rm -rf ~/$script_dir
fi

if [[ -e ~/$pkg_dir ]]; then
  echo "Removing installed packages"
  rm -rf ~/$pkg_dir
fi

# cleaning up the key storage in the server
echo "Removing from authorized_hosts"
ssh $user@$ip "rm ~/.ssh/authorized_keys; touch ~/.ssh/authorized_keys"

if [[ -e ~/.ssh ]]; then
  echo "Removing SSH folder"
  rm -rf ~/.ssh
fi

