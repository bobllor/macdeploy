#!/bin/bash

source ../globals.sh

if [[ $(fdesetup isactive) == false ]]; then
  echo "Starting FileVault process"
  logger $log_file "Starting FileVault process"
  
  # probably not needed since we full wipe, but the user
  # is the one allowed to turn off filevault
  key_line=$(sudo -p "$pw_prompt" fdesetup enable -user $(whoami) -verbose 2>&1 /dev/tty | grep -vi fdesetup)

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
