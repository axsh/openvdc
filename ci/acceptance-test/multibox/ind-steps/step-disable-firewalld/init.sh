#!/bin/bash

(
    $starting_step "Disable firewalld"
    sudo chroot "${TMP_ROOT}" /bin/bash -c "systemctl is-enabled firewalld"
    [[ ! $? -eq 0 ]]
    $skip_step_if_already_done; set -xe
    sudo chroot "${TMP_ROOT}" /bin/bash -c "systemctl disable firewalld"
) ; prev_cmd_failed
