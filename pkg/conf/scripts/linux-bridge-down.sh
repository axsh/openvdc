#!/bin/sh

CONTAINER=$1
SECTION=$2 # net
ACTION=$3  # up/down
IFTYPE=$4  # veth,macvlan,etc...
IFNAME=$5

brctl delif {{.BridgeName}} "${IFNAME}"
