#!/bin/bash

# Enables the Firewall of the device.

# i am unsure how this interaction works with go, but i dont think
# much is changed.
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate on