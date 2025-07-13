#!/bin/bash

# constant variables used for the script
# WARNING: DO NOT MODIFY THIS UNLESS YOU KNOW WHAT YOU ARE DOING
# if there are any issues contact Tri before making changes.

#################
## SERVER VARS ##
#################

user=donotmigrate
ip=10.142.46.165 # private ip of the server

# ssh only
ssh_deploy=$user@ip://Users/donotmigrate/mac-deployment
# if the file path is needed, ssh only
deploy=/Users/donotmigrate/mac-deployment

pkg_dir=pkg-files # directory of the pkg files for the software
script_dir=client-files # directory of the scripts for the deployment

support_dir=support-scripts # directory of the support scripts for deployment
