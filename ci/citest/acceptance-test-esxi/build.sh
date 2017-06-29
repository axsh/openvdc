#!/bin/bash

TMP_IP=192.168.2.103
NEW_IP=192.168.2.110 #Todo: Assign this somehow
DISKSPACE="20GB"
MEMORY="8192"
NETWORK="VM Network"
ISO_DATASTORE="datastore1"
ISO="images/centos7ks2.iso"
VM_DATASTORE="datastore1"
OS="centos7_64Guest"

IP_ADDR=$TMP_IP

function run_ssh() {
    local key=ssh_key
    [[ -f ${key} ]] &&
        $(type -P ssh) -i "${key}" -o 'StrictHostKeyChecking=no' -o 'LogLevel=quiet' -o 'UserKnownHostsFile /dev/null' "${@}"
}

function ssh_cmd() {
	local cmd="$1"
	run_ssh ${VMUSER}@$IP_ADDR "$cmd"
}

function wait_for_vm_to_boot () {
  echo "Waiting for VM to boot up..."
  max_attempts=100
  while ! govc guest.start -l=${VMUSER}:${VMPASS} -vm=$VMNAME /bin/echo "Test"  2> /dev/null ; do
          sleep 5
          attempt=$(( attempt + 1 ))
          [[ $attempt -eq ${max_attempts} ]] && exit 1
  done
  echo "VM has now booted up."
}

function get_device_name () {
  ssh_cmd `ip -o link show | awk -F': ' '{print$2}' | grep -Fvx -e lo`
}

function assign_ip () {
  DEVICE=$(get_device_name)
  ssh_cmd "sed '/^IPADDR/s/.*/IPADDR=${NEW_IP}/' /etc/sysconfig/network-scripts/ifcfg-${DEVICE} >> /etc/sysconfig/network-scripts/ifcfg-tmp"
  ssh_cmd "mv -f /etc/sysconfig/network-scripts/ifcfg-tmp /etc/sysconfig/network-scripts/ifcfg-${DEVICE}"
}

function build_vm () {
  govc vm.create -ds="$VM_DATASTORE" -iso="$ISO" -iso-datastore="$ISO_DATASTORE" -net="$NETWORK" -g="$OS" -disk="$DISKSPACE" -m="$MEMORY" -on=true -dump=true $VMNAME
  echo "Sent command: vm.create -ds=$VM_DATASTORE -iso=$ISO -iso-datastore=$ISO_DATASTORE -net=$NETWORK -g=$OS -disk=$DISKSPACE -m=$MEMORY -on=true -dump=true $VMNAME"
  
  wait_for_vm_to_boot

  add_ssh_key
  sleep 5
  ssh_cmd "yum install -y http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm"
  ssh_cmd "yum install -y mesos"
  ssh_cmd "yum install -y mesosphere-zookeeper"
  ssh_cmd "systemctl disable mesos-slave"
  ssh_cmd "systemctl disable mesos-master"
  assign_ip
  ssh_cmd "shutdown -h 0"
  sleep 5
  IP_ADDR=$NEW_IP

  echo "Saving VM ${VMNAME} > ${BACKUPNAME}"
  ovftool -ds=$VM_DATASTORE -n="$BACKUPNAME" --noImageFiles $FIXED_URL$VMNAME $FIXED_URL
}

function add_ssh_key () {
  rm -f ./ssh_key
  rm -f ./ssh_key.pub
  ssh-keygen -t rsa -b 2048 -N "" -f ./ssh_key
  chmod 600 ./ssh_key.pub
  govc guest.mkdir -l "${VMUSER}:${VMPASS}" -vm="$VMNAME" -p /root/.ssh
  govc guest.upload -l "${VMUSER}:${VMPASS}" -vm="$VMNAME" ./ssh_key.pub /root/.ssh/authorized_keys
  [[ -f ~/.ssh/known_hosts ]] && ssh-keygen -R $IP_ADDR
  #TODO: Copy new ssh-keys to cache
}

function install_yum_package () {
  local package="$1"  
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
    export GOVC_DATACENTER='ha-datacenter'
  fi

  if [[ -z "${GOVC_INSECURE}" ]] ; then
    export GOVC_INSECURE=true
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

TRIMMED_URL=$(echo $GOVC_URL | tr -d ' sdk')
FIXED_URL=$(sed 's/http/vi/g' <<< $TRIMMED_URL)

YUM_REPO_URL="https://ci.openvdc.org/repos/${BRANCH}/${RELEASE_SUFFIX}/"
curl -fs --head "${YUM_REPO_URL}" > /dev/null
if [[ "$?" != "0" ]]; then
  echo "Unable to reach '${YUM_REPO_URL}'."
  echo "Are the BRANCH and RELEASE_SUFFIX set correctly?"
  exit 1
fi

if [[ "$REBUILD" == "true" ]]; then
  if [[ $(govc vm.info $VMNAME -A) ]]; then
    echo "Old VM found. Attempting to delete it."
    govc guest.rm $VMNAME
  fi
  if [[ $(govc vm.info $BACKUPNAME -A) ]]; then
    echo "Old Backup found. Attempting to delete it."
    govc guest.rm $BACKUPNAME
  fi

  build_vm

else
  if [[ $(govc vm.info $VMNAME -A) ]]; then
    echo "Old VM found. Attempting to delete it."
    govc guest.rm $VMNAME
  fi

  if [[ $(govc vm.info $BACKUPNAME -A) ]]; then
      echo "Creating VM. ${BACKUPNAME} > ${VMNAME} "
      ovftool -ds=$VM_DATASTORE -n="$VMNAME" --noImageFiles $FIXED_URL$BACKUPNAME $FIXED_URL
    else
      echo "${BACKUPNAME} not found. Building VM:"
      build_vm
    fi
fi

govc vm.power -on=true $VMNAME

echo "Getting VM IP..."
IP_ADDR=$(govc vm.ip $VMNAME)

ssh_cmd "cat > /etc/yum.repos.d/openvdc.repo << EOS
[openvdc]
name=OpenVDC
failovermethod=priority
baseurl=${YUM_REPO_URL}
enabled=1
gpgcheck=0
EOS"

ssh_cmd "yum install -y openvdc"
ssh_cmd "systemctl enable openvdc-scheduler"
ssh_cmd "systemctl start openvdc-scheduler"
ssh_cmd "systemctl enable mesos-slave"
ssh_cmd "systemctl start mesos-slave"

ssh_cmd "systemctl enable mesos-master"
ssh_cmd "systemctl start mesos-master"
