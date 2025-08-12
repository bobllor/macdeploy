#!/bin/bash

# script used to enable filevault.
#
# only run this script via the go program. do not run this normally.
# args:
#   - $0: the username of the admin account.
#   - $1: the password of the admin account.

user=$0
pw=$1

plistVal="
<plist>
    <dict>
        <key>Username</key>
        <string>$user</string>
        <key>Password</key>
        <string>$pw</string>
    </dict>
</plist>
"

sudo fdesetup enable -inputplist <<< $plistVal