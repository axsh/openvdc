#!/bin/bash

(
    $starting_step "Create the public key and setup ssh config for ${user}"
    [[ -f ${CACHE_DIR}/${BRANCH}/sshkey_${vm_name} ]]
    $skip_step_if_already_done; set -ex
    add_user_key "${ci_user}"
    mv ${NODE_DIR}/sshkey ${CACHE_DIR}/${BRANCH}/sshkey_${vm_name}
) ; prev_cmd_failed


(
    $starting_step "Install authorized ssh key for ${user} on ${vm_name}"
    sudo bash -c "[ -f ${TMP_ROOT}/${user}/.ssh/authorized_keys ]"
    $skip_step_if_already_done; set -ex
    install_user_key "${ci_user}"
) ; prev_cmd_failed

# Quick hack for now. Should be changed in the seed image instead
(
    $starting_step "Disable DNS on sshd"
    grep -qw "UseDNS yes" "${TMP_ROOT}/etc/ssh/sshd_config"
    $skip_step_if_already_done; set -ex
    sed -i 's/UseDNS yes/UseDNS no/g' "${TMP_ROOT}/etc/ssh/sshd_config"
) ; prev_cmd_failed
