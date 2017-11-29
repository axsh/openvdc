package esxi

import "github.com/axsh/openvdc/model"

type Nic struct {
	IfName       string
	Index        string
	Ipv4Addr     string
	MacAddr      string
	Bridge       string
	BridgeHelper string
	Type         string
}

type baseImage struct {
	name string
	datastore string
}

type EsxiMachine struct {
	baseImage         baseImage
	SerialConsolePort int
	Nics              []Nic
}

func newEsxiMachine(serialPort int, template *model.EsxiTemplate_Image_Template) *EsxiMachine {
	return &EsxiMachine{
		SerialConsolePort: serialPort,
		baseImage: baseImage{
			name: template.GetName(),
			datastore: template.GetDatastore(),
		},
	}
}

func (m *EsxiMachine) AddNICs(nics []Nic) {
	for _, nic := range nics {
		m.Nics = append(m.Nics, nic)
	}
}
