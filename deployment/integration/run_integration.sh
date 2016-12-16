#!/bin/bash

set -e

BUILD_ENV_PATH=${1:?"ERROR: env file is not given."}

if [[ -n "${BUILD_ENV_PATH}" && ! -f "${BUILD_ENV_PATH}" ]]; then
  echo "ERROR: Can't find the file: ${BUILD_ENV_PATH}" >&2
  exit 1
fi

# http://stackoverflow.com/questions/19331497/set-environment-variables-from-file
set -a
. ${BUILD_ENV_PATH}
set +a


CID=openvdc-integration    # In future -- add a suffix to allow for concurrency
function vbox_rm() {
    if [[ -z "$CID" ]]; then
        return 0
    fi
    if [[ -n "$LEAVE_VM" ]]; then
        if [[ "${LEAVE_VM}" != "0" ]]; then
            echo "Skip to clean container: ${CID}"
            return 0
        fi
    fi

    ssh yumrepo@192.168.56.111  "vboxmanage controlvm  ${CID} poweroff   && vboxmanage unregistervm  ${CID}  --delete"
}

trap "vbox_rm  "  EXIT

cp /var/lib/jenkins/.ssh/id_rsa.pub deployment/integration
echo ${RELEASE_SUFFIX} > ./release_id
tar cf - . |  ssh yumrepo@192.168.56.111 tar xf - -C /data/openvdc-integration
ssh yumrepo@192.168.56.111  "cd /data/openvdc-integration/deployment/integration &&  ./build.sh "


ssh yumrepo@192.168.56.111  "cd /data/openvdc-integration/deployment/integration/output-virtualbox-ovf && vboxmanage import openvdc-integration.ovf && vboxmanage startvm ${CID}  --type headless"

started=$(date '+%s')
while ! (echo "" | nc 192.168.56.61 22) > /dev/null; do
  echo "Waiting for 192.168.56.61:sshd starts to listen ..."
  sleep 1
  if [[ $(($started + 60)) -le $(date '+%s') ]]; then
    echo "Timed out for 192.168.56.61:sshd becomes ready"
    exit 1
  fi
done

tar cf - . |  ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no vagrant@192.168.56.61 tar xf - -C /var/tmp --warning=no-timestamp  --no-overwrite-dir


## Here is the actual integration test code
