#!/bin/bash

set -e

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
tar cf - . |  ssh yumrepo@192.168.56.111 tar xf - -C /data/openvdc-integration 
ssh yumrepo@192.168.56.111  "cd /data/openvdc-integration/deployment/integration &&  ./build.sh"


ssh yumrepo@192.168.56.111  "cd /data/openvdc-integration/deployment/integration/output-virtualbox-ovf && vboxmanage import openvdc-integration.ovf && vboxmanage startvm ${CID}  --type headless"

tar cf - . |  ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no vagrant@192.168.56.61 tar xf - -C /var/tmp --warning=no-timestamp  --no-overwrite-dir


## Here is the actual integration test code
