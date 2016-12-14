#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/ind-steps/common.source"

copy_default_config
. "${ENV_ROOTDIR}/config.source"

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

for node in ${scheduled_nodes[@]} ; do
    (
        $starting_group "Destroy ${node%,*}"
        false
        $skip_group_if_unnecessary
        branch="${BRANCH}" ${ENV_ROOTDIR}/${node}/destroy.sh
    ) ; prev_cmd_failed
done

[[ -z "${1}" ]] || exit 1

function destroy_bridge() {
  local name="$1"

  (
    $starting_step "Destroy bridge ${name}"
    brctl show | grep -q "${name}"
    [ "$?" != "0" ]
    $skip_step_if_already_done ; set -xe
    sudo ip link set "${1}" down
    sudo brctl delbr "${1}"
  ) ; prev_cmd_failed
}

(
    $starting_group "Destroy bridges"
    false
    $skip_group_if_unnecessary
    destroy_bridge "vdc_env_br0"
) ; prev_cmd_failed

stop_masquerade "${NETWORK}/${PREFIX}"
