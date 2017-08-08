#!/bin/bash

# set -x

: "${BACKUP_PATH:?"BACKUP_PATH should be set"}" 
: "${ESXI_IP:?"ESXI_IP should be set"}" 
: "${ESXI_KEY}:?"ESXI_KEY should be set"}" 
: "${ESXI_USER:?"ESXI_USER should be set"}"

BACKUP_PATH=$"${BACKUP_PATH}/$(date +%Y%m%d_%H%M%S)"
BACKUP_OBJECTS=(
    "datastore1/images"
    "datastore2/base"
    "datastore2/CentOS7"
    "datastore2/images"
    "datastore2/key"
)

ssh_esxi() {
    ssh -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}" "${@}"
}
scp_esxi() {
    scp_esxi() {
        local remote_path="${1}"

        if $(ssh_esxi "bash [ -f ${remote_path} ]"); then
            scp -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}:${remote_path}" "${BACKUP_PATH}/${2}"
        elif $(ssh_esxi "bash [ -d ${remote_path} ]"); then
            scp -r -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}:${remote_path}" "${BACKUP_PATH}/${2}"
        else
            echo "Unable to find remote file, ${remote_path}, exiting"
            exit 1
        fi
}

# backup config files
mkdir -p "${BACKUP_PATH}"
ssh_esxi "vim-cmd hostsvc/firmware/sync_config"
ssh_esxi "vim-cmd hostsvc/firmware/backup_config"
backup_file="$(ssh_esxi find . \| grep configBundle-vmware.tgz)"
scp_esxi "${backup_file#*.}" "${backup_file##*/}"
ssh_esxi "rm -r ${backup_file}"

# backup datastore

# TODO: filter out the requried images/look into possible tools that will do this automatically
datastore_volumes=( "$(ssh_esxi ls -la /vmfs/volumes \| grep datastore)" )
datastore_id=()

for token in ${datastore_volumes[@]} ; do
    [[ "${next_is_id}" == "true" ]] && {
        datastore_id+=( "${token}" )
        mkdir -p "${BACKUP_PATH}/${name}/${token}"
        unset next_is_id
        unset name
        continue
    }
    [[ "${token}" == *"->"* ]] && next_is_id=true
    [[ "${token}" == *"datastore"* ]] && name="${token}"
done

for bo in ${BACKUP_OBJECTS[@]} ; do
    ds="${bo%/*}"
    id="${ds#*datastore}"
    # scp_esxi "/vmfs/volumes/${bo}" "${datastore_id[$(( id - 1 ))]}"
done

rm -f ${BACKUP_PATH%/*}/latest
ln -s ${BACKUP_PATH} ${BACKUP_PATH%/*}/latest
