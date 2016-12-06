
(
    $starting_step "Prepare openvdc repo"
    [[ -f "${TMP_ROOT}/etc/yum.repos.d/openvdc.repo" ]]
    $skip_step__if_already_done ; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -e <<EOS
    cat <<EOF >> /etc/yum.repos.d/openvdc.repo
[openvdc]
name=OpenVDc Repo - devrepo
baseurl=http://ci.openvdc.org/repos/20161205100450git44bd030/
enabled=1
gpgcheck=0
EOF
EOS
) ; prev_cmd_failed


