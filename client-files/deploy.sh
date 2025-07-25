#!/bin/bash

# created by: Tri Nguyen
# before you run this script, ensure it is on the CLIENT computer rather than the server.
# this is intended to be ran on the client computer due to the ssh requirement.

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
		ssh-copy-id donotmigrate@10.142.46.165
	fi
			
	cd ~/client-files

	source ./utils.sh
	source ./globals.sh
	source ./logging.sh
	source ./clean_up.sh

	cd ~

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
	logger $log_file "Install TeamViewer: $1 | Admin: $2"

	# this will take some time due to the file sizes.
	if [[ ! -e ~/$pkg_dir ]]; then
		echo "Installing required packages..."
		scp -r $user@$ip://Users/$user/mac-deployment/$pkg_dir ~ && logger $log_file "Successfully installed package folders"
	fi
	
	# NOTE: there could be a better way to do this
	#if [[ $(pkgutil --pkg-info com.apple.pkg.RosettaUpdateAuto) =~ "No receipt" ]]; then
	sudo softwareupdate --install-rosetta --agree-to-license
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
		sudo installer -pkg $(readlink -f $line) -target / \
		&& logger $log_file "Installed $pkg_name" \
		|| logger $log_file "Failed to install $pkg_name" 3; else \
		echo "$pkg_name already installed"; 
		logger $log_file "Skipping installation of $pkg_name, already installed"; 
		fi;
		sleep .5
	done

	sleep .3	

	# user creation
	echo "Creating user accounts"
	logger $log_file "Starting user account creation process"
	
	# addUser is the home directory, fullName is the display name
	sudo sysadminctl -addUser help.account -fullName "Help.Account" -password "@ssistMe"

	echo ""
	echo "Valid name formats: First Last | F Last | First.Last | F.Last"
	echo "The input is case sensitive, it follows the format above."
	echo -n "Enter the name of the user (enter nothing to skip): "
	read name_input

	echo ""
	
	if [[ -n "$name_input" ]]; then
		name_regex="^([A-Z]|[A-Z]([a-z]+))( |\.)[A-Z]([a-z]+)$"
		
		# replaces any spaces with periods
		username=$(sed "s/ /./" <<< $name_input)
	
		if [[ $username =~ $name_regex ]]; then
			# following macOS naming convention, their folder is lowercase
			account_dir=$(echo $username | awk '{print tolower()}')
			
			if [[ $2 == 'false' ]]; then
				sudo sysadminctl -addUser "$account_dir" -fullName "$username" -password "Password1" 
			else
				sudo sysadminctl -addUser "$account_dir" -fullName "$username" -password "Password1" -admin
			fi
			
			echo "Created user $name_input"
			logger $log_file "User $name_input created, username: $username | dir name: $account_dir"
		else
			echo "Invalid name read, manual input is required"
			logger $log_file "Invalid name input was given: $name_input" 3
		fi	
	else
		echo "No name given, skipping user creation process"
		logger $log_file "Skipped user creation"
	fi

	# FIXME: need to handle potential errors in this
	if [[ $(fdesetup isactive) == false ]]; then
		echo "Starting FileVault process"
		logger $log_file "Starting FileVault process"
		
		# probably not needed since we full wipe, but the user
		# is the one allowed to turn off filevault
		key_line=$(sudo fdesetup enable -user $(whoami) -verbose 2>&1 /dev/tty | grep -vi fdesetup)

		# if the stdout is an error then this will not run	
		if [[ "$(echo $key_line | grep -qi "error"; echo $?)" == 1 ]]; then
			# extract the key from the line
			key=$(echo $key_line | cut -d "'" -f2)
			  
			serial=$(get_serial)
		  
			logger $log_file "Generated $key for $serial"
			
			# NOTE: this needs to be changed if the path is changed  
			server_key=$(ssh $user@$ip "bash ~/mac-deployment/server-scripts/key_checker.sh '$serial' '$key'" | xargs echo)

			logger $log_file "$server_key"
			
			echo "Generated $key for $serial, stored in server"

			echo "FileVault activated"
			logger $log_file "FileVault activated"
		else
			echo "Something went wrong with FileVault,  manual interaction required"
			logger $log_file "FileVault error: $key_line" 3
		fi
	else
		echo "FileVault already enabled"
		logger $log_file "FileVault already enabled" 
	fi

	logger $log_file "Deployment process finished for $(get_serial)"
	echo "Moving generated log to server"
	move_log $log_file

	clean_up $script_dir $pkg_dir	
	
	# i need to research this further, the firewall command exits the script early.
	# i moved it down to the last step to avoid issues later.
	# this could be because i am accessing the socket directly instead of running a command?
	if [[ $(/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate | grep -qi "enabled"; echo $?) == 1 ]]; then
		echo "Enabling Firewall"
		sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate on

		echo "Firewall enabled"
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