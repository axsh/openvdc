
(
    $starting_step "Install mesos"
    sudo chroot ${TMP_ROOT} /bin/bash -c "rpm -q mesos" > /dev/null
    $skip_step_if_already_done ; set -xe
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y mesos"
    $master || sudo chroot ${TMP_ROOT} /bin/bash -c "systemctl disable mesos-master"
    $slave || sudo chroot ${TMP_ROOT} /bin/bash -c "systemctl disable mesos-slave"
) ; prev_cmd_failed
