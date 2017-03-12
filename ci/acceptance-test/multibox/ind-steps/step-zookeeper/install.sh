(
    $starting_step "Install zookeeper on ${vm_name}"
    sudo chroot ${TMP_ROOT} /bin/bash -c "rpm -q mesos-zookeeper" > /dev/null
    $skip_step_if_already_done; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y mesosphere-zookeeper"
    $zk_host || sudo chroot ${TMP_ROOT} /bin/bash -c "systemctl disable zookeeper"
) ; prev_cmd_failed
