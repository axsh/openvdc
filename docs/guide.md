

## Start a LXC instance with primitive commands.

```
% openvdc register centos/7/lxc
r-0000001
% openvdc create r-0000001
i-0000001
% openvdc start i-0000001
% openvdc show i-0000001
{
  "instance_id": "i-0000001",
  "slave_id": "xxxxx",
  "vcpu": 1,
  "memory_gb": 1
}
% openvdc stop i-0000001
% openvdc destroy i-0000001
```

## Setup Juniper SSG

```
% openvdc register juniper/ssg16 '{"console_ip":"192.168.1.10/24"}'
r-0000001
% openvdc create r-0000001
ssg-0000001
% openvdc start ssg-000001
% openvdc console ssg-000001 < myconfig.junos
configure JUNOS
% openvdc destroy ssg-000001
```
