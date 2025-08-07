#!/bin/bash

# argument is required due to this being embedded.
user=$0

# the output is captured by the exec call and will have parsing logic in that aspect.
# the grep is necessary to only output out the final key
sudo fdesetup enable -user $user -verbose 2>&1 /dev/tty | grep -vi fdesetup