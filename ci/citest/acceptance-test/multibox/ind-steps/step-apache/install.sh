(
    $starting_step "Install Apache on ${vm_name}"
    sudo chroot ${TMP_ROOT} /bin/bash -c "rpm -q httpd" > /dev/null
    $skip_step_if_already_done; set -ex
    sudo chroot ${TMP_ROOT} /bin/bash -c "yum install httpd -y" 
    sudo chroot ${TMP_ROOT} /bin/bash -c "mkdir -p /var/www/html/images"
) ; prev_cmd_failed
