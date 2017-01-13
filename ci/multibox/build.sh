#!/bin/bash

# Set the PATH variable so chrooted centos will know where to find stuff
export PATH=/bin:/sbin:/usr/bin:/usr/sbin

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))" && pwd -P)"
. "${ENV_ROOTDIR}/ind-steps/common.source"
. "${ENV_ROOTDIR}/config.source"

# Jenkins writes its environment variables to a build.env file in OpenVDC root
if [[ -f "${ENV_ROOTDIR}/../../build.env" ]]; then
  . "${ENV_ROOTDIR}/../../build.env"
fi

require_branch_variable
require_rebuild_variable
#TODO: require release suffix

YUM_REPO_URL="https://ci.openvdc.org/repos/${BRANCH}/${RELEASE_SUFFIX}/"
curl -fs --head "${YUM_REPO_URL}" > /dev/null
if [[ "$?" != "0" ]]; then
  echo "Unable to reach '${YUM_REPO_URL}'."
  echo "Are the BRANCH and RELEASE_SUFFIX set correctly?"
  exit 1
fi

export BRANCH
export REBUILD
export RELEASE_SUFFIX
export YUM_REPO_URL

scheduled_nodes=${NODES[@]}
[[ -n "$1" ]] && scheduled_nodes="${@}"

check_dep "tar"
check_dep "rsync"
check_dep "brctl"
check_dep "qemu-system-x86_64"
check_dep "parted" # For mount-partition.sh

(
  $starting_step "Enable IP forwarding"
  [[ "$(cat /proc/sys/net/ipv4/ip_forward)" == "1" ]]
  $skip_step_if_already_done
  sudo sysctl -w net.ipv4.ip_forward=1
) ; prev_cmd_failed

for box in ${BOXES} ; do
    download_seed_image "${box}"
done

create_bridge "vdc_env_br0" "${GATEWAY}/${PREFIX}"

if [[ "$REBUILD" == "true" ]]; then
    (
        $starting_group "Cleanup old environment"
        [ ! -d "${CACHE_DIR}/${BRANCH}" ]
        $skip_group_if_unnecessary
        rm -rf "${CACHE_DIR}/${BRANCH}"
        for node in ${scheduled_nodes[@]} ; do
            (
                $starting_group "Destroying ${node%,*}"
                false
                $skip_group_if_unnecessary
                "${ENV_ROOTDIR}/${node}/destroy.sh"
            ) ; prev_cmd_failed
        done
    ) ; prev_cmd_failed
fi

(
    $starting_step "Create cache folder"
    [ -d "${CACHE_DIR}/${BRANCH}" ]
    $skip_step_if_already_done ; set -ex
    mkdir -p "${CACHE_DIR}/${BRANCH}"
) ; prev_cmd_failed

masquerade "${NETWORK}/${PREFIX}"

for node in ${scheduled_nodes[@]} ; do
    (
        $starting_group "Building ${node%,*}"
        false
        $skip_group_if_unnecessary ; set -xe
        "${ENV_ROOTDIR}/${node}/build.sh"
    ) ; prev_cmd_failed
done
