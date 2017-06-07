package qemu

import (
	"github.com/pkg/errors"
)
type NetDev struct {
	IfName       string
	Id           string
	Type         string
	MacAddr      string
	Bridge       string
	BridgeHelper string
}

func NewNetworkDevice(t string, id string) (*NetDev, error) {
	if t != "tap" {
		return &NetDev{},errors.Errorf("Currently only type 'tap' is supported")
	}

	return &NetDev{
		Id: id,
		Type: t,
	},nil
}

func (nd *NetDev) SetBridge(brName string) error {
	nd.Bridge = brName
	return nil
}

func (nd *NetDev) SetName(name string) error {
	nd.IfName = name
	return nil
}

func (nd *NetDev) SetMacAddr(addr string) error {
	nd.MacAddr = addr
	return nil
}
