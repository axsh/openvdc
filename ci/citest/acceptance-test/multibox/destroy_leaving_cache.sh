#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/external_libraries/bashsteps/simple-defaults-for-bashsteps.source"
. "${ENV_ROOTDIR}/ind-steps/common.source"

. "${ENV_ROOTDIR}/config.source"

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

for node in ${scheduled_nodes[@]} ; do
    ${ENV_ROOTDIR}/${node}/destroy.sh
done

(
  $starting_step "Destroy simulated global interface"
  ip link | grep -q "${GLOBAL_TAP}"
  [ "$?" != "0" ]
  $skip_step_if_already_done ; set -xe
  sudo ip link set "${GLOBAL_TAP}" down
  sudo ip link delete dev "${GLOBAL_TAP}"
) ; prev_cmd_failed

destroy_bridge "vdc_mngnt"
destroy_bridge "vdc_insts"
stop_masquerade "${NETWORK}/${PREFIX}"
