## Install

```
yum install -y http://openvdc.org/openvdc-release.rpm
yum install -y openvdc
yum install -y http://openvnet.org/openvnet.repo
yum install -y openvnet
```

## Setup


```
systemctl enable zookeeper
systemctl enable mesos-master
systemctl enable mesos-slave
systemctl enable openvdc-api
systemctl enable openvdc-executor
systemctl start openvdc-api
systemctl start openvdc-executor
systemctl enable mysql-server
systemctl enable vnet-vnmgr
systemctl enable vnet-vna
systemctl enable redis
systemctl enable openvswitch
```

Setup basic network for OpenVNet.

```
vnctl networks add \
  --uuid nw-test1 \
  --display-name testnet1 \
  --ipv4-network 10.100.0.0 \
  --ipv4-prefix 24 \
  --network-mode virtual
vnctl interfaces add \
  --uuid if-inst1 \
  --mode vif \
  --owner-datapath-uuid dp-test1 \
  --mac-address 10:54:ff:00:00:01 \
  --network-uuid nw-test1 \
  --ipv4-address 10.100.0.10 \
  --port-name inst1
```


## Start an instance with virtual network

```
% openvdc run centos/7 --bridge[0]=localovs1 --interface[0]=10:54:ff:00:00:01 --ipv4[0]=10.100.0.10/24 --ipv4-gw[0]=10.100.0.1
abdcefg12345678
% openvdc ssh user@abdcefg12345678
% openvdc destroy abdcefg12345678
```

```
% openvdc run centos/7 --networkmode=openvnet --bridge[0]=nw-test1 --interface[0]=if-inst1
abdcefg12345678
% openvdc ssh user@abdcefg12345678
% openvdc destroy abdcefg12345678
```
