// +build linux

package qemu

import (
	// "fmt"
	"os"
	"path/filepath"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
)

type BridgeType int

const (
	None BridgeType = iota // 0
	Linux
	OVS
)

func (t BridgeType) String() string {
	switch t {
	case Linux:
		return "linux"
	case OVS:
		return "ovs"
	default:
		return "none"
	}
}

type QEMUHypervisorProvider struct {
}

type QEMUHypervisorDriver struct {
	hypervisor.Base
	template  *model.QemuTemplate
	imageName string
	hostName  string
	machine   *qemu.Machine
}

func (p *QEMUHypervisorProvider) Name () string {
	return "qemu"
}

var settings struct {
	ImgageServerUri string
	CachePath       string
	BridgeType      BridgeType
	BridgeName      string
	InstancePath    string
}

func init() {
	hypervisor.RegisterProvider("qemu", &QEMUHypervisorProvider{})
	viper.SetDefault("hypervisor.image-server-uri", "http://127.0.0.1/images")
	viper.SetDefault("hypervisor.cache-path", "/var/cache/qemu")
	viper.SetDefault("hypervisor.instance-path", "/var/openvdc/qemu-instances")
}

func (p *QEMUHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	if sub.IsSet("bridge.name") {
		setting.BridgeName = sub.GetString("bridge.name")
		if sub.IsSet("bridge.type") {
			switch sub.GetString("bridge.type") {
			case "linux":
				settings.BridgeType = Linux
			case "ovs"
				settings.BridgeType = OVS
			default:
				return errors.Errorf("Unknown bridges.type value: %s". sub.GetString("bridges.type"))
			}
		}
	} else if sub.IsSet("bridges.linux.name") {
n		log.Warn("bridges.linux.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.linux.name")
		settings.BridgeType = Linux
	} else if sub.IsSet("bridges.ovs.name") {
		log.Warn("bridges.ovs.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.ovs.name")
		settings.BridgeType = OVS
	}

	u := sub.GetString("hypervisor.image-server-uri")
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return errors.Errorf("Error parsing hypervisor.image-server-uri: %s", u)
	}

	settings.ImageServerUri = u
	settings.CachePath = sub.GetString("hypervisor.cache-path")
	settings.InstancePath = sub.GetString("hypervisor.instance-path")
	return nil
}

func (p *QEMUHypervisorProvider) CreateDriver (instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	qemuTmpl, ok := template.(*model.QemuTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.QemuTemplate: %T, template")
	}
	m := qemu.NewMachine(int(qemuTmpl.Vcpu), uint64(qemuTmpl.MemoryGb))
	driver := &QEMUHypervisorDriver{
		Base: hypervisor.Base{
			Log: log.WithFields(log.Fields{"Hypervisor": "qemu", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: qemuTmpl,
		machine: m,
	}
	return driver, nil
}

func (d *QEMUHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func (d *QEMUHypervisorDriver) getImage() {
}

func (d *QEMUHypervisorDriver) buildMachine(imagePath string) error {
	d.machine.AddDrive(qemu.Drive{
		Path: imagePath,
		Format: d.format,
	})

	var netDev []qemu.NetDev

	netDev = append(netDev, qemu.NetDev{
		Type: "tap",
		Id: "test1",
		IfName: fmt.Sprintf("testif"),
		MacAddr: "00:00:00:00:00:02",
	})

	d.machine.AddNICs(netDev)
	return nil
}

func (d *QEMUHypervisorDriver) CreateInstance() error {
	instanceId := d.template.Base.instance.GetId()
	instanceDir := filepath.Join(settings.InstancePath, instanceId)
	imagePath := filepath.Join(instanceDir, "diskImage."+d.template.Image.Format)

	os.MkdirAll(instanceDir, os.ModePerm)
	if _, err := os.Stat(baseImage) ; err != nil {
		d.getImage()
	}
	if _, err := os.Stat(imagePath) ; err != nil {
		img, _ := qemu.NewImage(imagePath, d.format, baseImage)
		img.CreateInstanceImage()
	}

	d.buildMachine(imagePath)
	return nil
}

func (d *QEMUHypervisorDriver) DestroyInstance() error {
	return nil
}

func (d *QEMUHypervisorDriver) StartInstance() error {
	return nil
}

func (d *QEMUHypervisorDriver) StopInstance() error {
	return nil
}

func (d QEMUHypervisorDriver) RebootInstance() error {
	return nil
}

func (d QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return nil
}
