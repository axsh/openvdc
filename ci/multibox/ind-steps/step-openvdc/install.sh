(
    $starting_step "Install openvdc on ${vm_name}"

    false
    $skip_step_if_already_done; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y openvdc --skip-broken"
) ; prev_cmd_failed
