#!/bin/bash

# utils for the scripts

script_dir=client-files

source ~/$script_dir/globals.sh

get_serial(){
	serial_line=$(ioreg -l | grep IOPlatformSerialNumber) 
	
	# this may require a change with f4 to something else if
	# mac decides to change this for whatever reason.
	echo "$serial_line" | cut -d '"' -f4
}	

# checks for installed applications, used to stop a reinstall
check_installed(){
	string=$(echo "$1" | awk '{print tolower($0)}')

	if [[ $string =~ "teamview" ]]; then
		val=$(ls /Applications/ | grep -Ei "teamview")
		
		test -n "$val" && echo 0 || echo 1
	elif [[ $string =~ "absolute" ]]; then
		val=$(ls /Library/Application\ Support/ | grep -i "absolute")
			
		test -n "$val" && echo 0 || echo 1
	elif [[ $string =~ "office" ]]; then
		val=$(ls /Applications/ | grep -Ei "(word|excel|outlook|powerpoint)")

		test -n "$val" && echo 0 || echo 1
	fi	
}
