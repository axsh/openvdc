package qemu

type cmdLine struct {
	args []string
}

func (c *cmdLine) appendArgs(args... string) {
	for _, arg := range args {
		c.args = append(c.args, arg)
	}
}

func (cmd *cmdLine) qemuBootCmd(m *Machine, useKvm bool) []string {
	cmd.appendArgs("-smp", strconv.Itoa(m.Cores), "-m", strconv.FormatUint(m.Memory, 10))
	if useKvm {
		cmd.appendArgs("-enable-kvm")
	}
	if len(m.Serial) > 0 {
		cmd.appendArgs("-serial", m.Serial)
	}
	if len(m.Monitor) > 0 {
		cmd.appendArgs("-monitor", m.Monitor)
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
			cmd.appendArgs("-netdev", fmt.Sprintf("%s,ifname=%s,script=%s,downscript=%s,id=%s",
				nic.Type, nic.IfName, nic.Upscript, nic.Downscript, nic.Id))
			device := fmt.Sprintf("virtio-net-pci,netdev=%s", nic.Id)
			if len(nic.MacAddr) > 0 {
				device = fmt.Sprintf("%s,mac=%s", device, nic.MacAddr)
			}
			cmd.appendArgs("-device", device)
		}
	}
	return cmd.args
}

func (cmd *cmdLine) qemuImgCmd(i *Image) []string {
	cmd.appendArgs("create", "-f", i.Format, "-b", i.baseImg, i.Path)
	return cmd.args
}
