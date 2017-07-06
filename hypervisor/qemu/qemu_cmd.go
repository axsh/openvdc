package qemu

import (
	"fmt"
	"strconv"
)

type cmdLine struct {
	args []string
}

func (c *cmdLine) appendArgs(args ...string) {
	for _, arg := range args {
		c.args = append(c.args, arg)
	}
}

func (cmd *cmdLine) QemuBootCmd(m *Machine) []string {
	cmd.appendArgs("-smp", strconv.Itoa(m.Cores), "-m", strconv.FormatUint(m.Memory, 10))
	if m.Kvm {
		cmd.appendArgs("-enable-kvm")
	}
	if len(m.SerialSocketPath) > 0 {
		cmd.appendArgs("-serial", fmt.Sprintf("unix:%s,server,nowait", m.SerialSocketPath))
	}
	if len(m.MonitorSocketPath) > 0 {
		cmd.appendArgs("-monitor", fmt.Sprintf("unix:%s,server,nowait", m.MonitorSocketPath))
	}
	if len(m.Vnc) > 0 {
		cmd.appendArgs("-vnc", m.Vnc)
	}
	if len(m.Pidfile) > 0 {
		cmd.appendArgs("-pidfile", m.Pidfile)
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
	cmd.appendArgs("-daemonize")
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
