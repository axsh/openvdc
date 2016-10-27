# Getting started


## Install

```
yum install -y http://openvdc.org/openvdc-release.rpm
yum install -y openvdc
```

## Setup???


```
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
