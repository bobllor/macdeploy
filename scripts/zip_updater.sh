#!/usr/bin/env bash

# Accesses the endpoint that triggers the ZIP update API.
# This should only be used by the container.

cd /macdeploy/src/server
log="/tmp/cronner.log"
touch $log

timer_arg=$(awk '{print tolower($0)}' <<< $1)
sec_to_hour=3600

# default 2 hours
timer=7200
reg='^[0-9]+[mhs]$'
if [[ "$timer_arg" =~ $reg ]]; then
    if [[ "$timer_arg" =~ "h" ]]; then
        val=$(sed s/h//g <<< $timer_arg)
        timer=$(($val * $sec_to_hour))
    elif [[ "$timer_arg" =~ "m" ]]; then
        val=$(sed s/m//g <<< $timer_arg)
        timer=$(($val * 60))
    else
        timer=$timer_arg
    fi
fi

echo "Timer: $timer | Arg: $timer_arg" >> $log

# Not ideal... but I cannot figure out how to start the cron service as root
# and change to a non-root user in the container.
while true; do
    sleep $timer

    log_word_count=$(cat $log | wc -l)
    if [[ $log_word_count -gt 100 ]]; then
        log_file_count=$(ls /tmp | grep -E "cronner\.*" | wc -l)
        log="/tmp/cronner-$log_file_count.log"
    fi

    token=$(cat .token)
    host="python-fs:5000" # this must match the name of the compose.yml name.

    curl -H "x-zip-token: $token" https://$host/api/zip/update --insecure \
        &>> $log
done