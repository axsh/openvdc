#!/bin/bash

IP=$1

function get_device_name () {
  ip -o link show | awk -F': ' '{print$2}' | grep -Fvx -e lo
}

function assign_ip () {
  DEVICE=$(get_device_name)
  sed '/^IPADDR/s/.*/IPADDR='${IP}'/' /etc/sysconfig/network-scripts/ifcfg-${DEVICE} >> /etc/sysconfig/network-scripts/ifcfg-tmp
  mv -f /etc/sysconfig/network-scripts/ifcfg-tmp /etc/sysconfig/network-scripts/ifcfg-${DEVICE}
}

assign_ip
