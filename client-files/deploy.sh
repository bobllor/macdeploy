#!/bin/bash

# created by: Tri Nguyen
# main deployment script for macOS
#
# IMPORANT:
#   the main branch has a CI/CD script into it. as much as i want to add testing,
#   unfortunately i made the decision to write this in bash only.
#   git pull the main branch on a new branch and test it properly.
#   fuck it we ball.

get_date(){
	echo $(date +"%m-%d-%yT%H-%M-%S")
}

init(){
	# prevents accidental usage on the server (would be very bad)		
	if [[ $(whoami) == donotmigrate ||  -e ~/.ssh/authorized_keys ]]; then
		echo "ERROR: Script ran on the server, exiting"
		exit 1
	fi

	# this is used as a backup check in case the one-liner fails
	if [[ ! -e ~/.ssh/ ]]; then
		# will be using rsa by default
		ssh-keygen -f ~/.ssh/id_rsa -N ""
    echo "Enter password to the SERVER."
		ssh-copy-id donotmigrate@10.142.46.165
	fi

	cd ~/client-files

	source ./utils.sh
	source ./globals.sh
	source ./logging.sh

	# used to validate for files on the server	
	ssh $user@$ip "bash ~/mac-deployment/server-scripts/validation.sh"
}			

# args:
# 	$1 (-T): Default false, indicates if TeamViewer should be installed.
# 	$2 (-A): Default false, indicates if the user should be Admin.
main(){
	init

	log_file="$(get_serial)-$(get_date)".log
	logger $log_file "Starting deployment for $(get_serial)"
	logger $log_file "Install TeamViewer: $1 | Admin: $2" 1

	if [[ ! -e ~/$pkg_dir ]]; then
		echo "Installing required packages..."
		scp -r $user@$ip://Users/$user/mac-deployment/$pkg_dir ~ && \ 
      logger $log_file "Successfully installed package folders" || \
      echo "CRITICAL: FAILED TO INSTALL SCRIPTS"; exit 1 # probably not needed but won't be bad to have
	fi

	# NOTE: there could be a better way to do this
	#if [[ $(pkgutil --pkg-info com.apple.pkg.RosettaUpdateAuto) =~ "No receipt" ]]; then
	sudo -p "$pw_prompt" softwareupdate --install-rosetta --agree-to-license
	#fi

	regex="(full|teamviewer|office)"
	
	if [[ $1 == 'false' ]]; then
		regex="(full|office)"
	fi
  
	find ~/$pkg_dir -type f -name "*.pkg" \
	| grep -Ei "$regex" \
	| while read -r line; do \
		pkg_name=$(basename $line);
		if [[ $(check_installed "$line") == 1 ]]; then \
		sudo -p "$pw_prompt" installer -pkg $(readlink -f $line) -target / \
		&& logger $log_file "Installed $pkg_name" \
		|| logger $log_file "Failed to install $pkg_name" 3; else \
		echo "$pkg_name already installed"; 
		logger $log_file "Skipping installation of $pkg_name, already installed"; 
		fi;
		sleep .5
	done

	sleep .3	

	# user creation
  bash ./$support_dir/user_creation.sh $log_file

  bash ./$support_dir/filevault_activation.sh $log_file

	logger $log_file "Deployment process finished for $(get_serial)"
	echo "Moving generated log to server"
	move_log $log_file

	bash ./$support_dir/clean_up.sh	

  # do not move this anywhere else, this must be the last execution
	if [[ $(/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate | grep -qi "enabled"; echo $?) == 1 ]]; then
    bash ./$support_dir/firewall.sh
	else
		echo "Firewall already enabled"
	fi
}

tv_status='false'
admin='false'

while getopts 'TA' flag; do
	case "${flag}" in
		T) tv_status='true' ;; 
		A) admin='true' ;;
		*) 
			echo "Invalid option read"
			exit 1;;
	esac
done

main $tv_status $admin
