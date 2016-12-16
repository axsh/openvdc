#!/bin/bash

# Set the PATH variable so chrooted centos will know where to find stuff
export PATH=/bin:/sbin:/usr/bin:/usr/sbin

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/ind-steps/common.source"

copy_default_config
. "${ENV_ROOTDIR}/config.source"
export BRANCH

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

check_dep "wget"
check_dep "tar"
check_dep "rsync"
check_dep "brctl"
check_dep "qemu-system-x86_64"
check_dep "parted" # For mount-partition.sh

for box in ${BOXES} ; do
    download_seed_image "${box}"
done

create_bridge "vdc_env_br0" "${GATEWAY}/${PREFIX}"

$REBUILD && {
    (
        $starting_group "Cleanup old environment"
        sudo [ ! -d "${CACHE_DIR}/${BRANCH}" ]
        $skip_group_if_unnecessary
        sudo rm -rf "${CACHE_DIR}/${BRANCH}"
        for node in ${scheduled_nodes[@]} ; do
            (
                $starting_group "Destroying ${node%,*}"
                false
                $skip_group_if_unnecessary
                "${ENV_ROOTDIR}/${node}/destroy.sh"
            ) ; prev_cmd_failed
        done
    ) ; prev_cmd_failed

    (
        $starting_step "Create cache folder"
            sudo [ -d "${CACHE_DIR}/${BRANCH}" ]
            $skip_step_if_already_done ; set -ex
            sudo mkdir -p "${CACHE_DIR}/${BRANCH}"
    ) ; prev_cmd_failed
} || {
    (
        $starting_step "Clone base images from ${BASE_BRANCH}"
        [ -d "${CACHE_DIR}/${BRANCH}" ]
        $skip_step_if_already_done ; set -ex
        sudo cp -r "${CACHE_DIR}/${BASE_BRANCH}" "${CACHE_DIR}/${BRANCH}"
    ) ; prev_cmd_failed
}

masquerade "${NETWORK}/${PREFIX}"

for node in ${scheduled_nodes[@]} ; do
    (
        $starting_group "Building ${node%,*}"
        false
        $skip_group_if_unnecessary
        "${ENV_ROOTDIR}/${node}/build.sh"
    ) ; prev_cmd_failed
done



