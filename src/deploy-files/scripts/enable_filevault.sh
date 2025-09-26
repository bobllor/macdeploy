#!/bin/bash

# Enable FileVault with a plist for automation.
#
# Args:
#   - $0: The username of the admin account.
#   - $1: The password of the admin account.

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