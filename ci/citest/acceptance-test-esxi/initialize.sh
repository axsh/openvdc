#!/bin/bash

: "${BACKUP_PATH:?"BACKUP_PATH should be set"}" 
: "${ESXI_IP:?"ESXI_IP should be set"}" 
: "${ESXI_KEY}:?"ESXI_KEY should be set"}" 
: "${ESXI_USER:?"ESXI_USER should be set"}"

BACKUP_PATH=$"${BACKUP_PATH}/latest"
ssh_esxi() {
    ssh -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}" "${@}"
}
scp_esxi() {
    local local_path="${BACKUP_PATH}/${1}"
    if [[ -f "${local_path}" ]]; then
        scp -i "${ESXI_KEY}" "${BACKUP_PATH}/${1}" "${ESXI_USER}@${ESXI_IP}:${1}"
    elif [[ -d "${local_path}" ]]; then
        scp -i "${ESXI_KEY}" "${BACKUP_PATH}/${1}" "${ESXI_USER}@${ESXI_IP}:${1}"
    else
        echo "Unable to find source file, ${local_path}, exiting"
        exit 1
    fi
}

[[ -d "${BACKUP_PATH}" ]] || {
    echo "Unable to find backup source, exiting..."
    exit 1
}

ssh_esxi "vim-cmd hostsvc/maintenance_mode_enter"
scp_esxi "configBundle-vmware.tgz" "/tmp/configBundle.tgz"
ssh_esxi "vim-cmd hostsvc/firmware/restore_config /tmp/configBundle.tgz"
for ds in $(ls -d ${BACKUP_PATH}/latest/*) ; do
    [[ -d "${ds}" ]] && {
        id="$(ls ${ds})"
        echo scp_esxi "${ds}/${id}" "/vmfs/volumes/"
        echo ssh_esxi "ln -s /vmfs/volumes/${id} /vmfs/volumes/${ds##*/}"
    }
done
