#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/ind-steps/common.source"
. "${ENV_ROOTDIR}/config.source"

require_branch_variable

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

for node in ${scheduled_nodes[@]} ; do
    ${ENV_ROOTDIR}/${node}/destroy.sh
    ${ENV_ROOTDIR}/${node}/destroy_cache.sh
done

destroy_bridge "vdc_mngnt"
destroy_bridge "vdc_insts"
stop_masquerade "${NETWORK}/${PREFIX}"
stop_masquerade "${NETWORK_INSTS}/${PREFIX}"

(
  $starting_step "Remove cache directory"
  [[ ! -d "${CACHE_DIR}/${BRANCH}" ]]
  $skip_step_if_already_done; set -xe
  rmdir "${CACHE_DIR}/${BRANCH}"
) ; $prev_cmd_failed
