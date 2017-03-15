#!/bin/bash

(
    $starting_step "Append entries to /etc/hosts"
    grep "10.0.100." ${TMP_ROOT}/etc/hosts > /dev/null
    $skip_step_if_already_done; set -xe
    cat <<EOF | sudo chroot "${TMP_ROOT}" /bin/bash -c 'cat >> /etc/hosts'
10.0.100.10 zookeeper
10.0.100.11 mesos
10.0.100.12 scheduler
10.0.100.13 executor-null
10.0.100.14 executor-lxc
EOF
) ; prev_cmd_failed
