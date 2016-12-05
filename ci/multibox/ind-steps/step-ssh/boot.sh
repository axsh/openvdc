#!/bin/bash

(
    $starting_step "Wait for ssh"
    [[ "$(nc ${IP_ADDR} 22 < /dev/null)" == *"SSH"* ]]
    $skip_step_if_already_done ; set -xe
    timeout=15
    while ! ssh root@${IP_ADDR} "uptime" > /dev/null ; do
        sleep 5
        tries=$(( tries + 1 ))
        [[ $tries -eq ${timeout} ]] && exit 1
    done
    :
) ; prev_cmd_failed

ssh-keygen -R ${IP_ADDR}
