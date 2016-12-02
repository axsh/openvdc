#!/bin/bash

(
    $starting_step "Create the public key and setup ssh config"
    [[ -f ${NODE_DIR}/root@${vm_name} ]]
    $skip_step_if_already_done; set -ex
    ssh-keygen -t rsa -b 2048 -N "" -f ${NODE_DIR}/root@${vm_name}
    chmod 600 ${NODE_DIR}/root@${vm_name}
    chmod 700 ${NODE_DIR}/root@${vm_name}.pub
) ; prev_cmd_failed

(
    $starting_step "Install authorized ssh key for ${vm_name}"
    [[ -f "${TMP_ROOT}/root/.ssh/authorized_keys" ]]
    $skip_step_if_already_done; set -xe
    sudo chroot ${TMP_ROOT} /bin/bash -ex <<EOS
mkdir -p -m 700 /root/.ssh

sed -i \
-e 's,^PermitRootLogin .*,PermitRootLogin yes,' \
-e 's,^PasswordAuthentication .*,PasswordAuthentication yes,' \
-e 's,^DenyUsers.root,#DenyUsers root,' \
\
/etc/ssh/sshd_config
EOS
    sudo cp "${NODE_DIR}/root@${vm_name}.pub" "${TMP_ROOT}/root/.ssh/authorized_keys"

) ; prev_cmd_failed


