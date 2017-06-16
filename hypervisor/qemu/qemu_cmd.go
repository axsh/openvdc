package qemu

import (
	"fmt"
	"strconv"
)
type cmdLine struct {
	args []string
}

func (c *cmdLine) appendArgs(args... string) {
	for _, arg := range args {
		c.args = append(c.args, arg)
	}
}

func (cmd *cmdLine) QemuBootCmd(m *Machine) []string {
	cmd.appendArgs("-smp", strconv.Itoa(m.Cores), "-m", strconv.FormatUint(m.Memory, 10))
	if m.Kvm {
		cmd.appendArgs("-enable-kvm")
	}
	if len(m.Serial) > 0 {
		cmd.appendArgs("-serial", fmt.Sprintf("unix:%s,server,nowait",m.Serial))
	}
	if len(m.Monitor) > 0 {
		cmd.appendArgs("-monitor", fmt.Sprintf("unix:%s,server,nowait", m.Monitor))
	}
	if len(m.Vnc) > 0 {
		cmd.appendArgs("-vnc", m.Vnc)
	}
	for _, d := range m.Drives {
		drive := fmt.Sprintf("file=%s,format=%s", d.Image.Path, d.Image.Format)
		if len(d.If) > 0 {
			drive = fmt.Sprintf("%s,if=%s", drive, d.If)
		}
		cmd.appendArgs("-drive", drive)
	}
	if len(m.Nics) == 0 {
		cmd.appendArgs("-net", "none")
	} else {
		for _, nic := range m.Nics {
			// brdev := fmt.Sprintf("bridge,br=%s", nic.Bridge)
			// if len(nic.BridgeHelper) > 0 {
			// 	brdev = fmt.Sprintf("%s,helper=%s", brdev, nic.BridgeHelper)
			// }
			// netdev := fmt.Sprintf("nic,model=virtio")
			// if len(nic.MacAddr) > 0 {
			// 	netdev = fmt.Sprintf("%s,macaddr=%s", netdev, nic.MacAddr)
			// }
			// cmd.appendArgs("-net", brdev)
			// cmd.appendArgs("-net", netdev)

			netdevOpts := fmt.Sprintf("tap,ifname=%s,id=%s", nic.IfName, nic.IfName)
			deviceOpts := fmt.Sprintf("virtio-net-pci,netdev=%s", nic.IfName)
			if len(nic.MacAddr) > 0 {
				deviceOpts = fmt.Sprintf("%s,mac=%s", deviceOpts, nic.MacAddr)
			}

			cmd.appendArgs("-netdev", netdevOpts)
			cmd.appendArgs("-device", deviceOpts)
		}
	}
	cmd.appendArgs("-display", m.Display)
	return cmd.args
}

func (cmd *cmdLine) QemuImgCmd(i *Image) []string {
	cmd.appendArgs("create", "-f", i.Format)
	cmd.appendArgs(i.Path)

	if len(i.baseImage) > 0 {
		cmd.appendArgs("-b", i.baseImage)
	} else {
		cmd.appendArgs(fmt.Sprintf("%sK", strconv.Itoa(i.Size)))
	}
	return cmd.args
}
