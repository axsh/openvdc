
(
    $starting_step "Install mesos"
    sudo chroot ${TMP_ROOT} /bin/bash -c "which mesos-master" > /dev/null
    $skip_step_if_already_done ; set -xe
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y mesos"
) ; prev_cmd_failed
