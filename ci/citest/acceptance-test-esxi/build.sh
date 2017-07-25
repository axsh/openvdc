#!/bin/bash

set -ex -o pipefail

BUILD_ENV_PATH=${1:?"ERROR: env file is not given."}

if [[ -n "${BUILD_ENV_PATH}" && ! -f "${BUILD_ENV_PATH}" ]]; then
  echo "ERROR: Can't find the file: ${BUILD_ENV_PATH}" >&2
  exit 1
fi

set -a
. ${BUILD_ENV_PATH}
set +a

IP_ADDR=$IP_ADDRESS
NETWORK="VM Network"

function run_ssh() {
  local key=ssh_key
  [[ -f ${key} ]] &&
  $(type -P ssh) -i "${key}" -o 'StrictHostKeyChecking=no' -o 'LogLevel=quiet' -o 'UserKnownHostsFile /dev/null' "${@}"
}

function ssh_cmd() {
  local cmd="${@}"
  run_ssh ${VMUSER}@$IP_ADDR "$cmd"
}

function clone_base_vm () {
  BASE="base"
  echo "Cloning VM. ${BASE} > ${1}"
  set +x;
  echo "yes" | ovftool -ds=$VM_DATASTORE -n="$1" --noImageFiles $2$BASE $2; set -x;

  govc vm.power -on=true $VMNAME

  echo "Waiting for VM to get assigned IP"
  govc vm.ip -wait 3m $VMNAME
  sleep 10

  add_ssh_key

  yum_install "http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm"
  yum_install "mesos"
  yum_install "mesosphere-zookeeper"
  ssh_cmd "/bin/systemctl disable mesos-slave"
  ssh_cmd "/bin/systemctl disable mesos-master"

  ssh_cmd "shutdown -h 0"
  sleep 30

  echo "Saving VM ${VMNAME} > ${BACKUPNAME}"
  set +x;
  echo "yes" | ovftool -ds=$VM_DATASTORE -n="$BACKUPNAME" --noImageFiles $FIXED_URL$VMNAME $FIXED_URL; set -x+
}

function add_ssh_key () {
  rm -f ./ssh_key
  rm -f ./ssh_key.pub
  ssh-keygen -t rsa -b 2048 -N "" -f ./ssh_key
  chmod 600 ./ssh_key.pub
  govc guest.mkdir -l "${VMUSER}:${VMPASS}" -vm="$VMNAME" -p /root/.ssh
  govc guest.upload -l "${VMUSER}:${VMPASS}" -vm="$VMNAME" ./ssh_key.pub /root/.ssh/authorized_keys
  ssh-keygen -R $IP_ADDR
}

function vm_cmd () {
  PSID=$(govc guest.start -l=${VMUSER}:${VMPASS} -vm ${VMNAME} $@)
  govc guest.ps -l=${VMUSER}:${VMPASS} -vm ${VMNAME} -p $PSID
}

function yum_install () {
  local package="$1"
  ssh_cmd "yum install -y ${package}"
}

function check_dep() {
  local dep="$1"

  command -v "${dep}" >/dev/null 2>&1
  if [[ "$?" != "0" ]]; then
    echo "Missing dependency: ${dep}"
    exit 1
  fi
}

function check_env_variables () {
  if [[ -z "${IP_ADDRESS}" ]] ; then
    echo "The IP_ADDRESS variable needs to be set."
    exit 1
  fi
  if [[ -z "${NETWORK}" ]] ; then
    echo "The NETWORK variable needs to be set."
    exit 1
  fi
  if [[ -z "${VM_DATASTORE}" ]] ; then
    echo "The VM_DATASTORE variable needs to be set."
    exit 1
  fi
  if [[ -z "${BOX_DISKSPACE}" ]] ; then
    echo "The BOX_DISKSPACE variable needs to be set."
    exit 1
  fi
  if [[ -z "${BOX_MEMORY}" ]] ; then
    echo "The BOX_MEMORY variable needs to be set."
    exit 1
  fi
  if [[ -z "${VMUSER}" ]] ; then
    echo "The VMUSER variable needs to be set."
    exit 1
  fi

  if [[ -z "${VMPASS}" ]] ; then
    echo "The VMPASS variable needs to be set."
    exit 1
  fi

  if [[ -z "${GOVC_URL}" ]] ; then
    echo "The GOVC_URL variable needs to be set. Example: https://username:password@ip/sdk"
    exit 1
  fi

  if [[ -z "${GOVC_DATACENTER}" ]] ; then
    echo "The GOVC_DATACENTER variable needs to be set."
    exit 1
  fi

  if [[ -z "${GOVC_INSECURE}" ]] ; then
    echo "The GOVC_INSECURE variable needs to be set."
    exit 1
  fi

  if [[ -z "${BRANCH}" ]] ; then
    echo "the BRANCH variable needs to be set with the github branch to test."
    exit 1
  fi

  if [[ -z "${REBUILD}" ]] ; then
    echo "The REBUILD variable needs to be set. 'true' if you wish to rebuild the environment completely. 'false' otherwise"
    exit 1
  fi

  if [[ -z "${RELEASE_SUFFIX}" ]] ; then
    echo "the RELEASE_SUFFIX variable needs to be set with the release suffix in the yum repo we're testing. Usually looks similar to: '20170111063228git2d0dc08'."
    exit 1
  fi
}

check_dep "ssh"
check_dep "govc"
check_dep "ovftool"
check_env_variables

VMNAME="$BRANCH"
BACKUPNAME="${VMNAME}_BACKUP"
set +x;
TRIMMED_URL=$(echo $GOVC_URL | tr -d ' sdk')
FIXED_URL=$(sed 's/http/vi/g' <<< $TRIMMED_URL)
set -x;
YUM_REPO_URL="https://ci.openvdc.org/repos/${BRANCH}/${RELEASE_SUFFIX}/"
curl -fs --head "${YUM_REPO_URL}" > /dev/null
if [[ "$?" != "0" ]]; then
  echo "Unable to reach '${YUM_REPO_URL}'."
  echo "Are the BRANCH and RELEASE_SUFFIX set correctly?"
  exit 1
fi

if [[ "$REBUILD" == "true" ]]; then
  if [[ $(govc vm.info $VMNAME) ]]; then
    echo "Old VM found. Attempting to delete it."
    govc vm.destroy $VMNAME
  fi
  if [[ $(govc vm.info $BACKUPNAME) ]]; then
    echo "Old Backup found. Attempting to delete it."
    govc vm.destroy $BACKUPNAME
  fi
  build_vm
else
  if [[ $(govc vm.info $VMNAME) ]]; then
    echo "Old VM found. Attempting to delete it."
    govc vm.destroy $VMNAME
  fi

  if [[ $(govc vm.info $BACKUPNAME) ]]; then
      echo "Creating VM. ${BACKUPNAME} > ${VMNAME} "
      echo "yes" | ovftool -ds=$VM_DATASTORE -n="$VMNAME" --noImageFiles $FIXED_URL$BACKUPNAME $FIXED_URL
    else
      echo "${BACKUPNAME} not found. Building VM:"
      clone_base_vm $VMNAME $FIXED_URL
    fi
fi

govc vm.power -on=true $VMNAME

echo "Waiting for VM to get assigned IP"
govc vm.ip -wait 3m $VMNAME
sleep 10

ssh_cmd "cat > /etc/yum.repos.d/openvdc.repo << EOS
[openvdc]
name=OpenVDC
failovermethod=priority
baseurl=${YUM_REPO_URL}
enabled=1
gpgcheck=0
EOS"

yum_install "openvdc"
ssh_cmd "systemctl enable openvdc-scheduler"
ssh_cmd "systemctl start openvdc-scheduler"
ssh_cmd "systemctl enable mesos-slave"
ssh_cmd "systemctl start mesos-slave"

ssh_cmd "systemctl enable mesos-master"
ssh_cmd "systemctl start mesos-master"
ssh_cmd "cp /opt/axsh/openvdc/bin/openvdc-executor /bin/"


echo "Installation complete."

# TODO: Run Tests

echo "Powering off VM..."
ssh_cmd "shutdown -h 0"
sleep 30

echo "Removing VM..."
govc vm.destroy $VMNAME
