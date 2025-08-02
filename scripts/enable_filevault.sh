#!/bin/bash

# the output is captured by the exec call and will have parsing logic in that aspect.
# the grep is necessary to only output out the final key
sudo -p "$pw_prompt" fdesetup enable -user $(whoami) -verbose 2>&1 /dev/tty | grep -vi fdesetup