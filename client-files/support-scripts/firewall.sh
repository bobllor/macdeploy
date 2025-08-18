#!/bin/bash

echo "Enabling Firewall"
sudo -p "$pw_prompt" /usr/libexec/ApplicationFirewall/socketfilterfw \ 
  --setglobalstate on

echo "Firewall enabled"
