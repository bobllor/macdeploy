#!/bin/bash

source ../logging.sh
source ../globals.sh

# needs to be passed due to the time stamp requirement
log_file=$1

echo "Creating user accounts"
logger $log_file "Starting user account creation process"

# addUser is the home directory, fullName is the display name
sudo -p "$pw_prompt" sysadminctl -addUser help.account \
  -fullName "Help.Account" -password "@ssistMe"

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
      sudo -p "$pw_prompt" sysadminctl -addUser "$account_dir" -fullName "$username" -password "Password1" 
    else
      sudo -p "$pw_prompt" sysadminctl -addUser "$account_dir" -fullName "$username" -password "Password1" -admin
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
