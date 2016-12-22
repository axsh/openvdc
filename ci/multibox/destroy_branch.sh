#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/ind-steps/common.source"

. "${ENV_ROOTDIR}/config.source"

kill_option="false"
[[ "$1" == "--kill" ]] && { kill_option="true" ; shift ; }

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

for node in ${scheduled_nodes[@]} ; do
    (
        $starting_group "Destroy ${node%,*}"
        false
        $skip_group_if_unnecessary
        if [[ "${kill_option}" == "true" ]]; then
          "${ENV_ROOTDIR}/${node}/kill.sh"
        else
          "${ENV_ROOTDIR}/${node}/destroy.sh"
        fi
    ) ; prev_cmd_failed
done

[[ -z "${1}" ]] || exit 1

destroy_bridge "vdc_env_br0"
stop_masquerade "${NETWORK}/${PREFIX}"
