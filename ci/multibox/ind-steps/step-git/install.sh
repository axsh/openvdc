
(
    $starting_step "Install git"
    [[ -f "${TMP_ROOT}/usr/bin/git" ]]
    $skip_step_if_already_done ; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y git"
) ; prev_cmd_failed
