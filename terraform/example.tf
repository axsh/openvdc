provider "openvdc" {
  api_endpoint = "0.0.0.0:5000"
}

resource "openvdc_instance" "joske" {
  template = "centos/7/lxc"

  interfaces = [
    {
      type = "veth" # default
      ipv4addr = "10.0.0.10" # optional
      macaddr = "10:54:ff:aa:00:04" # optional
      ipv4gateway = "10.0.0.1"
    },
    {
      type = "veth" # default
      ipv4addr = "10.0.1.10" # optional
      macaddr = "10:54:ff:aa:00:05" # optional
    },
  ]
}
