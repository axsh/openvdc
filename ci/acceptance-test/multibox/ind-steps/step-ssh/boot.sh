#!/bin/bash

(
    $starting_step "Import ssh key"
    [ -f ${NODE_DIR}/sshkey ]
    $skip_step_if_already_done
    cp ${CACHE_DIR}/${BRANCH}/sshkey_${vm_name} ${NODE_DIR}/sshkey
    chown ${USER}:${USER} ${NODE_DIR}/sshkey
) ; prev_cmd_failed

(
    $starting_step "Wait for ssh"
    [[ "$(nc ${IP_ADDR} 22 < /dev/null)" == *"SSH"* ]]
    $skip_step_if_already_done ; set -xe
    timeout=15
    while ! run_ssh root@${IP_ADDR} "uptime" > /dev/null ; do
        sleep 5
        tries=$(( tries + 1 ))
        [[ $tries -eq ${timeout} ]] && exit 1
    done
    :
) ; prev_cmd_failed

[[ -f ${USER}/.ssh/known_hosts ]] && ssh-keygen -R ${IP_ADDR}
