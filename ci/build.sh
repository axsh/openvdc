#!/bin/bash

# Set the PATH variable so chrooted centos will know where to find stuff
export PATH=/bin:/sbin:/usr/bin:/usr/sbin

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/ind-steps/common.source"

copy_default_config
. "${ENV_ROOTDIR}/config.source"

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

check_dep "wget"
check_dep "tar"
check_dep "rsync"
check_dep "brctl"
check_dep "qemu-system-x86_64"
check_dep "parted" # For mount-partition.sh

download_seed_image "${SEED_IMAGE}"
create_bridge "vdc_env_br0" "${GATEWAY}/${PREFIX}"

for node in ${scheduled_nodes[@]} ; do
    (
        $starting_group "Building ${node%,*}"
        sudo kill -0 $(sudo cat ${ENV_ROOTDIR}/${node}/${node}.pid 2> /dev/null) 2> /dev/null
        [[ "$?" == "0" ]]
        $skip_group_if_unnecessary
        "${ENV_ROOTDIR}/${node}/build.sh"
    ) ; prev_cmd_failed
done

masquerade "${NETWORK}/${PREFIX}"
