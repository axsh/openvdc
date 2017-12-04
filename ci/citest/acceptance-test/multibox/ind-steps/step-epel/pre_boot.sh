#!/bin/bash

(
    $starting_step "Install EPEL"
    sudo chroot "${TMP_ROOT}" /bin/bash -c \
        "rpm -qa epel-release* | egrep -q epel-release"
    [[ $? -eq 0 ]]
    $skip_step_if_already_done; set -xe
    # epel-release.rpm from CentOS/extra contains deprecated index for mirror sites.
    sudo chroot "${TMP_ROOT}" /bin/bash -e <<'EOF'
rpm -Uvh http://dl.fedoraproject.org/pub/epel/7/x86_64/Packages/e/epel-release-7-11.noarch.rpm
EOF

) ; prev_cmd_failed
