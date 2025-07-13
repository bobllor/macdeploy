#!/bin/bash

echo "Enabling Firewall"
sudo -p "$sudo_prompt" /usr/libexec/ApplicationFirewall/socketfilterfw \ 
  --setglobalstate on

echo "Firewall enabled"
