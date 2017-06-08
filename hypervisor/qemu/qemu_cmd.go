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

func (cmd *cmdLine) QemuBootCmd(m *Machine, useKvm bool) []string {
	cmd.appendArgs("-smp", strconv.Itoa(m.Cores), "-m", strconv.FormatUint(m.Memory, 10))
	if useKvm {
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
		cmd.appendArgs("-drive", fmt.Sprintf("file=%s,format=%s", d.Path, d.Format))
	}
	if len(m.Nics) == 0 {
		cmd.appendArgs("-net", "none")
	} else {
		for _, nic := range m.Nics {
			brdev := fmt.Sprintf("bridge,br=%s", nic.Bridge)
			if len(nic.BridgeHelper) > 0 {
				brdev = fmt.Sprintf("%s,helper=%s", brdev, nic.BridgeHelper)
			}
			netdev := fmt.Sprintf("nic,model=virtio")
			if len(nic.MacAddr) > 0 {
				netdev = fmt.Sprintf("%s,macaddr=%s", netdev, nic.MacAddr)
			}

			cmd.appendArgs("-net", brdev)
			cmd.appendArgs("-net", netdev)
		}
	}
	cmd.appendArgs("-display", m.Display)
	return cmd.args
}

func (cmd *cmdLine) QemuImgCmd(i *Image) []string {
	cmd.appendArgs("create", "-f", i.Format, "-b", i.baseImg, i.Path)
	return cmd.args
}
