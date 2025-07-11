#!/bin/bash

source ./globals.sh

clean_up(){
	rm ~/*.log

  script_dir=$1
  pkg_dir=$2

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
	
	# remove the key from the server
	echo "Removing from authorized_hosts"
	ssh $user@$ip "rm ~/.ssh/authorized_keys; touch ~/.ssh/authorized_keys"

	if [[ -e ~/.ssh ]]; then
		echo "Removing SSH folder"
		rm -rf ~/.ssh
	fi
}

