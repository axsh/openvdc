
(
    $starting_step "Download mesosphere repo"
    [[ -f "${TMP_ROOT}/etc/yum.repos.d/mesosphere.repo" ]]
    $skip_step_if_already_done ; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install -y http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm"
) ; prev_cmd_failed
