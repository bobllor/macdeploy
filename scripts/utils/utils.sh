#/usr/bin/env bash

# Get the value of an argument equalt to the variable name of the  configuration file at 
# src/server/configuration.py.
# Used to prevent hard coded strings in scripts by using configuration.py as the source of truth.
#
# Args:
#   - $1: The variable name.
filename(){
    var=$(python3 -c "from src.server.configuration import $1; print($1)" 2> /dev/null)

    printf "$var"
}

# Used to check if the filename returned a value or an empty string.
#
# Args:
#   - $1: The value of the filename function.
#   - $2: The variable name that was used in filename function.
varcheck(){
    if [[ -z "$1" ]]; then
        printf "Variable $2 could not be found, check $0 for argument name\n"
        return 1
    fi
}