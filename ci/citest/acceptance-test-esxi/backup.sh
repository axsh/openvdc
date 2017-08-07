#!/bin/bash

# set -x

: "${BACKUP_PATH:?"BACKUP_PATH should be set"}" 
: "${ESXI_IP:?"ESXI_IP should be set"}" 
: "${ESXI_KEY}:?"ESXI_KEY should be set"}" 
: "${ESXI_USER:?"ESXI_USER should be set"}"

ssh_esxi() {
    ssh -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}" "${@}"
}
scp_esxi() {
    scp -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}:${1}" "${BACKUP_PATH}/${2}"
}

# backup config files
ssh_esxi "vim-cmd hostsvc/firmware/sync_config"
ssh_esxi "vim-cmd hostsvc/firmware/backup_config"
backup_file="$(ssh_esxi find . \| grep configBundle-vmware.tgz)"
scp_esxi "${backup_file#*.}" "${backup_file##*/}_$(date +%Y%m%d)"
ssh_esxi "rm -r ${backup_file}"
ln -s "${BACKUP_PATH}/${backup_file##*/}_$(date +%Y%m%d)" "${BACKUP_PATH}/configBundle-vmware.tgz_latest"
# backup datastore

# TODO: filter out the requried images/look into possible tools that will do this automatically
datastore_volumes=( "$(ssh_esxi ls -la /vmfs/volumes \| grep datastore)" )
for token in ${datastore_volumes[@]} ; do
    [[ "${next_id}" == "true" ]] && {
        mkdir -p "${BACKUP_PATH}/$(date +%Y%m%d)/${name}"
        scp_esxi "/vmfs/volumes/${token}" "$(date +%Y%m%d)/${name}"
        ln -s "${BACKUP_PATH}/$(date +%Y%m%d)/${name}" "${BACKUP_PATH}/latest"

        unset next_is_id
        unset name
        continue
    }
    [[ "${token}" == *"->"* ]] && next_is_id=true
    [[ "${token}" == *"datastore"* ]] && name="${token}"
done
