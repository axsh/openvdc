
(
    $starting_step "Install gcc"
    [[ -f "${TMP_ROOT}/usr/bin/gcc" ]]
    $skip_step_if_already_done ; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y gcc"
) ; prev_cmd_failed
