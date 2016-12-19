#!/bin/bash

(
    $starting_step "Install LXC"
    [[ -f ${TMP_ROOT}/usr/bin/lxc-create ]]
    $skip_step_if_already_done; set -xe
    sudo chroot "${TMP_ROOT}" /bin/bash -c \
         "yum install -y lxc lxc-templates lxc-devel debootstrap"
) ; prev_cmd_failed
