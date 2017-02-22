resource "openvdc_instance" "joske" {
  template = "centos/7/lxc"

  interface {
    type = "veth" # default
    bridge = "ovs" # optional
    ipv4addr = "10.0.0.10" # optional
    macaddr = "10:54:ff:aa:00:04" # optional
  }
}
