#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))/.." && pwd -P)"
export NODE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "${NODE_DIR}/vmspec.conf"
. "${ENV_ROOTDIR}/ind-steps/step-box/common.source"

kill=${1}
$kill && kill-vm || destroy-vm
rm ${NODE_DIR}/root@${vm_name}*
