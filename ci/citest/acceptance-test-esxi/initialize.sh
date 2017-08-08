#!/bin/bash

: "${BACKUP_PATH:?"BACKUP_PATH should be set"}" 
: "${ESXI_IP:?"ESXI_IP should be set"}" 
: "${ESXI_KEY}:?"ESXI_KEY should be set"}" 
: "${ESXI_USER:?"ESXI_USER should be set"}"

ssh_esxi() {
    ssh -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}" "${@}"
}
scp_esxi() {
    scp -i "${ESXI_KEY}" "${BACKUP_PATH}/${2}" "${ESXI_USER}@${ESXI_IP}:${1}" 
}

ssh_esxi "vim-cmd hostsvc/maintenance_mode_enter"
scp_esxi "${BACKUP_PATH}/latest/configBundle-vmware.tgz" "/tmp/configBundle.tgz"
ssh_esxi "vim-cmd hostsvc/firmware/restore_config /tmp/configBundle.tgz"
for ds in $(ls -d ${BACKUP_PATH}/latest/*) ; do
    [[ -d "${ds}" ]] && {
        id="$(ls ${ds})"
        echo scp -r "${ds}/${id}" "/vmfs/volumes/"
        echo ssh_esxi "ln -s /vmfs/volumes/${id} /vmfs/volumes/${ds##*/}"
    }
done
