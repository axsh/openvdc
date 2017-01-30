# Resource Template

The resource template is a definition of datacenter resource.

Following is a LXC template as basic example:

```json
% cat ./templates/centos/7/lxc.json
{
  "title": "CentOS7",
  "template": {
    "type": "vm/lxc",
    "lxc_image": {
      "download_url": "https://images.linuxcontainers.org/1.0/images/d767cfe9a0df0b2213e28b39b61e8f79cb9b1e745eeed98c22bc5236f277309a/export"
    }
  }
}
```

You can check the syntax.

```bash
% openvdc template validate ./templates/centos/7/lxc.json
```

Show details about the template.

```bash
% openvdc template show ./templates/centos/7/lxc.json
lxc_image: <
  download_url: "https://images.linuxcontainers.org/1.0/images/d767cfe9a0df0b2213e28b39b61e8f79cb9b1e745eeed98c22bc5236f277309a/export"
>
```

Overwrite some parameters.

```bash
% openvdc template show ./templates/centos/7/lxc.json  '{"interfaces":[{"macaddr":"11:11:11:11:11:11"}]}'
lxc_image: <
  download_url: "https://images.linuxcontainers.org/1.0/images/d767cfe9a0df0b2213e28b39b61e8f79cb9b1e745eeed98c22bc5236f277309a/export"
>
interfaces: <
  macaddr: "11:11:11:11:11:11"
>
```
