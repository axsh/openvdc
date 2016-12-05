#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/external_libraries/bashsteps/simple-defaults-for-bashsteps.source"

copy_default_config
. "${ENV_ROOTDIR}/config.source"

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

for node in "${NODES[@]}" ; do
    ${ENV_ROOTDIR}/${node}/kill.sh ${kill_option}
done
