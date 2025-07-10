#!/bin/bash

script_dir=client-files

source ~/$script_dir/utils.sh
source ~/$script_dir/globals.sh

# not the same as get_date in deploy.
get_date_log(){
        echo $(date +"%Y-%m-%d %I:%M:%S %p")
}

logger(){
	msg=$2
	stat=INFO

	if [[ -z $msg ]]; then
		echo "No log message was given"
		exit 1
	fi
	if [[ $3 == 1 ]]; then
		stat=DEBUG
	elif [[ $3 == 3 ]]; then
		stat=WARNING
	elif [[ $3 == 4 ]]; then
		stat=ERROR
	elif [[ $3 == 5 ]]; then
		stat=CRITICAL
	fi

	echo "[$stat] $(get_date_log): $msg" >> ~/$1
}

move_log(){
	scp -q ~/$1 $user@$ip://Users/$user/mac-deployment/logs
}
