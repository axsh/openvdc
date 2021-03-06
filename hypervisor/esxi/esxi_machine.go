package esxi

import "github.com/axsh/openvdc/model"

type Nic struct {
	NetworkId   string
	IfName      string
	Index       string
	Ipv4Addr    string
	MacAddr     string
	Bridge      string
	Ipv4Gateway string
	Type        string
}

type baseImage struct {
	name      string
	datastore string
}

type EsxiMachine struct {
	baseImage         *baseImage
	SerialConsolePort int
	Nics              []Nic
}

func newEsxiMachine(serialPort int, template *model.EsxiTemplate) *EsxiMachine {
	base := template.GetEsxiImage()
	return &EsxiMachine{
		SerialConsolePort: serialPort,
		baseImage: &baseImage{
			name:      base.GetName(),
			datastore: base.GetDatastore(),
		},
	}
}
