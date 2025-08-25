#!/usr/bin/env bash

clear
set -e

temp_file_name="passcheck.txt"
temp_file="/tmp/$temp_file_name"

# used to check for errors during passwd command.
# this is a hacky but the exit status of passwd is always 0
error_file="/tmp/passerr.txt"

if [[ ! -e $temp_file ]]; then
	echo "false" > $temp_file
fi

# ensures removal of file so it doesn't error out on success
if [[ -e $error_file ]]; then
	rm $error_file
fi

has_completed_passwd=$(cat $temp_file)

if [[ $has_completed_passwd == "false" ]]; then
	echo "It is recommended to change your password on the device."
	echo "Please enter your current password, followed by a new password to use twice."
	echo -e "If there are failures in typing this password you can rerun this command file.\n"

	passwd 2> $error_file
	error_content=$(cat $error_file)
	
	if [[ ! -z $error_content ]]; then
		echo $error_content
		echo "Please restart the process again by clicking on the desktop file."
		exit 1
	fi
	
	echo "true" > $temp_file
fi

echo -e "Updating keychain for your account, please enter your old password and your new password\n"
security set-keychain-password

echo -e "\nSuccessfully updated password"

set +e

# debatable whether to keep this or not.
rm ~/Desktop/ChangePassword.command

# remove if successful, just in case this needs to be ran again.
rm $temp_file
rm $error_file