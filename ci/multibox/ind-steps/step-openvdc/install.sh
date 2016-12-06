(
    $starting_step "Install openvdc on ${vm_name}"
    [[ -f "${TMP_ROOT}/opt/axsh/openvdc/bin/openvdc" ]]
    $skip_step_if_already_done; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y openvdc"
) ; prev_cmd_failed
