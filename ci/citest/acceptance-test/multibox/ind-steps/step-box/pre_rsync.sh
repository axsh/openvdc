#!/bin/bash

if mount | grep -q "${TMP_ROOT}" ; then
    umount-seed-image
fi
(
    $starting_step "Deploy seed image for ${vm_name}"
    [ -f "$(vm_image)" ]
    $skip_step_if_already_done; set -xe
    tar -Sxzf "${BOXES_DIR}/${box}-${distr_ver}-${arch}.kvm.tar.gz" -C "${NODE_DIR}"
    rm ${NODE_DIR}/box-disk1.rpm-qa
) ; prev_cmd_failed

(
    $starting_step "Mount temporary root folder for ${vm_name}"
    mount | grep -q "${TMP_ROOT}"
    $skip_step_if_already_done
    mkdir -p "${TMP_ROOT}"
    mount-partition --sudo "$(vm_image)" 1 "${TMP_ROOT}"
) ; prev_cmd_failed
