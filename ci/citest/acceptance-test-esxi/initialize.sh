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
        scp -i "${ESXI_KEY}" "${BACKUP_PATH}/${1}" "${ESXI_USER}@${ESXI_IP}:${2}"
    elif [[ -d "${local_path}" ]]; then
        scp -r -i "${ESXI_KEY}" "${BACKUP_PATH}/${1}" "${ESXI_USER}@${ESXI_IP}:${2}"
    else
        echo "Unable to find source file: ${local_path}, exiting"
        exit 1
    fi
}

wait_for_ssh ()
{
    local timeout="${timeout:-300}"
    local sleep_time="${sleep_time:-10}"
    local readonly start_time=$(date +%s)

    while :; do
        ping -c 1 ${ESXI_IP > /dev/null && return 0
        [[ $(( $(date +%s) - start_time )) -gt $timeout ]] && return 0
        echo "Waiting for ssh..."
        sleep ${sleep_time}
    done

    echo "timed out, exiting..."
    exit 1
}

[[ -d "${BACKUP_PATH}" ]] || {
    echo "Unable to find backup source, exiting..."
    exit 1
}

scp_esxi "configBundle-vmware.tgz" "/tmp/configBundle.tgz"
ssh_esxi "vim-cmd hostsvc/maintenance_mode_enter"
ssh_esxi "vim-cmd hostsvc/firmware/restore_config /tmp/configBundle.tgz"

wait_for_ssh

for ds in $(ls -d ${BACKUP_PATH}/*) ; do
    [[ -d "${ds}" ]] && {
        id="$(ls ${ds})"
        echo scp_esxi "${ds}/${id}" "/vmfs/volumes/"
        echo ssh_esxi "ln -s /vmfs/volumes/${id} /vmfs/volumes/${ds##*/}"
    }
done
