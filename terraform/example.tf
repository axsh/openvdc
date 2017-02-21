resource "openvdc_instance" "joske" {
  template = "centos/7/lxc"
  options = '{"interfaces":[{"type":"veth", "bridge":"ovs"}]}'
}
