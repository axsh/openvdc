#!/bin/bash

export ENV_ROOTDIR="$(cd "$(dirname $(readlink -f "$0"))/.." && pwd -P)"
export NODE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP_ROOT="${NODE_DIR}/tmp_root"

. "${ENV_ROOTDIR}/config.source"
. "${NODE_DIR}/vmspec.conf"
. "${ENV_ROOTDIR}/ind-steps/common.source"

scheduler=true
zk_host=true

IND_STEPS=(
    "box"
    "ssh"
    "hosts"
    "disable-firewalld"
    "mesosphere-repo"
    "zookeeper"
)

build "${IND_STEPS[@]}"

# This is not part of the ind-steps because we don't want OpenVDC installed in
# the cached images. We want a clean cache without OpenVDC so we can install a
# different version to test every the CI runs.
install_openvdc_yum_repo
install_yum_package_over_ssh "openvdc-scheduler"
enable_service_over_ssh "openvdc-scheduler"
