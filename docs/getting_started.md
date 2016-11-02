# Getting started


## Install

```
yum install -y http://openvdc.org/openvdc-release.rpm
yum install -y openvdc
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
```


## Start an instance

```
% openvdc run centos/7
abdcefg12345678
% openvdc ssh abdcefg12345678
% openvdc list

% openvdc rm abdcefg12345678
```

```
% openvdc run centos/7 -name=myhost1
abdcefg12345678
% openvdc ssh myhost1
% openvdc list

% openvdc rm myhost1
```

## Start an instance with local bridge

```
% openvdc run centos/7 --bridge[0]=localovs1 --interface[0]=10:54:ff:00:00:01 --ipv4[0]=10.100.0.10/24 --ipv4-gw[0]=10.100.0.1
abdcefg12345678
% openvdc ssh user@abdcefg12345678
% openvdc rm abdcefg12345678
```
