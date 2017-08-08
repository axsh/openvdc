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
    local remote_path="${1}"
    # we can use recursive ssh copy here since it doesn't make a difference wether its a file or folder.
    scp -r -i "${ESXI_KEY}" "${ESXI_USER}@${ESXI_IP}:${remote_path}" "${BACKUP_PATH}/${2}"
    [[ $? == 1 ]] && {
        echo "Unable to find remote file: ${remote_path}, exiting"
        exit 1
    }
}

# backup config files
mkdir -p "${BACKUP_PATH}"
ssh_esxi "vim-cmd hostsvc/firmware/sync_config"
ssh_esxi "vim-cmd hostsvc/firmware/backup_config"
backup_file="$(ssh_esxi find . \| grep configBundle-vmware.tgz)"
scp_esxi "${backup_file#*.}" "${backup_file##*/}"
ssh_esxi "rm -r ${backup_file}"

# backup datastore
datastore_id=()
while read -r line ; do
    datastore_name=$(awk '{ print $9 }' <<< ${line})
    id=$(awk '{ print $11 }' <<< ${line})
    mkdir -p "${BACKUP_PATH}/${datastore_name}/${id}"
    datastore_id+=( "$id" )
done <<< "$(ssh_esxi ls -la /vmfs/volumes \| grep datastore)"

# ssh copy the files we want to backup and store them in their respective store id folder.
for bo in ${BACKUP_OBJECTS[@]} ; do
    ds="${bo%/*}"
    id="${ds#*datastore}"
    # datastore starts from index #1 while arrays start from 0
    scp_esxi "/vmfs/volumes/${bo}" "datastore$id/${datastore_id[$(( id - 1 ))]}"
done

rm -f ${BACKUP_PATH%/*}/latest
ln -s ${BACKUP_PATH} ${BACKUP_PATH%/*}/latest
