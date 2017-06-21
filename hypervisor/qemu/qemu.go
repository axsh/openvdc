// +build linux

package qemu

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"net/http"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/viper"
	"github.com/asaskevich/govalidator"
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
	machine   *Machine
}

func (p *QEMUHypervisorProvider) Name () string {
	return "qemu"
}

var settings struct {
	ImageServerUri  string
	CachePath        string
	BridgeType       BridgeType
	BridgeName       string
	InstancePath     string
	QemuBridgeHelper string
	QemuPath         string
	QemuProvider     string
}

func init() {
	hypervisor.RegisterProvider("qemu", &QEMUHypervisorProvider{})
	viper.SetDefault("hypervisor.image-server-uri", "http://127.0.0.1/images")
	viper.SetDefault("hypervisor.cache-path", "/var/cache/qemu")
	viper.SetDefault("hypervisor.instance-path", "/var/openvdc/qemu-instances")
}

func (p *QEMUHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	if _, err:= os.Stat("/usr/libexec/qemu-kvm"); err == nil {
		settings.QemuPath = "/usr/libexec"
		settings.QemuProvider = "qemu-kvm"
	} else if  _, err := os.Stat("/usr/bin/qemu-system-x86_64"); err == nil {
		settings.QemuPath = "/usr/bin"
		settings.QemuProvider = "qemu-system-x86_64"
	} else {
		return errors.Errorf("No qemu provider found.")
	}


	if sub.IsSet("bridge.name") {
		settings.BridgeName = sub.GetString("bridge.name")
		if sub.IsSet("bridge.type") {
			switch sub.GetString("bridge.type") {
			case "linux":
				settings.BridgeType = Linux
			case "ovs":
				settings.BridgeType = OVS
			default:
				return errors.Errorf("Unknown bridges.type value: %s", sub.GetString("bridges.type"))
			}
		}
	} else if sub.IsSet("bridges.linux.name") {
		log.Warn("bridges.linux.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.linux.name")
		settings.BridgeType = Linux
		settings.QemuBridgeHelper = filepath.Join(settings.QemuPath, "qemu-bridge-helper")
	} else if sub.IsSet("bridges.ovs.name") {
		log.Warn("bridges.ovs.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.ovs.name")
		settings.BridgeType = OVS
		settings.QemuBridgeHelper = "/path/to/qemu-ovs-helper"
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
	driver := &QEMUHypervisorDriver{
		Base: hypervisor.Base{
			Log: log.WithFields(log.Fields{"Hypervisor": "qemu", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: qemuTmpl,
	}
	driver.createMachineTemplate()
	return driver, nil
}

func (d *QEMUHypervisorDriver) createMachineTemplate ()  {
	instanceId := d.Base.Instance.GetId()
	instanceDir := filepath.Join(settings.InstancePath, instanceId)
	imageFormat := strings.ToLower(d.template.QemuImage.GetFormat().String())
	var netDev []NetDev
	for idx, iface := range d.template.GetInterfaces() {
		netDev = append(netDev, NetDev{
			IfName: fmt.Sprintf("%s_%02d", d.machine.Name, idx),
			Type: iface.Type,
			Ipv4Addr: iface.Ipv4Addr,
			MacAddr: iface.Macaddr,
			Bridge: settings.BridgeName,
			BridgeHelper: settings.QemuBridgeHelper,
		})
	}

	d.machine = NewMachine(int(d.template.GetVcpu()), uint64(d.template.GetMemoryGb()*1024))
	d.machine.Drives["disk"] = Drive{Image: NewImage(filepath.Join(instanceDir, "diskImage."+ imageFormat), imageFormat)}
	d.machine.Drives["meta"] = Drive{Image: NewImage(filepath.Join(instanceDir, "metadrive.img"), "raw"), If: "floppy"}
	d.machine.Name = instanceId
	d.machine.Monitor = fmt.Sprintf("%s",filepath.Join(instanceDir, "monitor.socket"))
	d.machine.Serial = fmt.Sprintf("%s",filepath.Join(instanceDir, "serial.socket"))
	d.machine.Kvm = d.template.GetUseKvm()
	d.machine.AddNICs(netDev)
}

func (d *QEMUHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func (d *QEMUHypervisorDriver) getImage() (string, error) {
	url := strings.Split(d.template.QemuImage.GetDownloadUrl(), "/")
	imageFile := url[len(url)-1]
	imageCachePath := filepath.Join(settings.CachePath, imageFile)

	if _, err := os.Stat(imageCachePath) ; err != nil {
		d.log().Infoln("Downloading machine image...")
		var remotePath string

		if govalidator.IsURL(d.template.QemuImage.GetDownloadUrl()) {
			remotePath = d.template.QemuImage.GetDownloadUrl()
		} else if settings.ImageServerUri != "" {
			remotePath = settings.ImageServerUri +"/"+ imageFile
		} else  {
			return "", errors.Errorf("Unable to resolve download_url: %s", d.template.QemuImage.GetDownloadUrl())
		}

		file, err := os.Create(imageCachePath)
		if err != nil {
			return "", errors.Errorf("Failed to create file: %s", imageCachePath)
		}
		resp, err := http.Get(remotePath)
		if err != nil {
			return "", errors.Errorf("Failed to download file: %s", remotePath)
		}
		defer resp.Body.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return "", errors.Errorf("Failed to download file: %s", remotePath)
		}
	}
	// todo check type if compressed type unpack and return unpacked filename
	return imageCachePath, nil
}

func runCmd(cmd string, args []string) error {
	c := exec.Command(cmd, args...)
	if err := c.Run(); err != nil {
		return errors.Errorf("failed to execute command :%s %s", cmd, args)
	}
	return nil
}

func (d *QEMUHypervisorDriver) addMetadata(metadataDrive *Image, datamap func(machine *Machine) map[string]string) error {
	mountPath := filepath.Join(filepath.Dir(metadataDrive.Path), "meta-data")

	os.MkdirAll(mountPath, os.ModePerm)
	if err := runCmd("mount", []string{metadataDrive.Path, mountPath}); err != nil {
		return errors.Errorf("Error: %s", err)
	}
	for key, value := range datamap(d.machine) {
		ioutil.WriteFile(filepath.Join(mountPath, key), []byte(value), 0644)
	}

	if err := runCmd("umount", []string{mountPath}); err != nil {
		return errors.Errorf("Error: %s", err)
	}
	os.RemoveAll(mountPath)
	return nil
}

func (d *QEMUHypervisorDriver) buildMetadriveBase(metadrive *Image) error {
	d.log().Infoln("Preparing metadrive image...")

	if err := runCmd("mkfs.msdos", []string{"-s", "1", metadrive.Path}); err != nil {
		return errors.Errorf("Error: %s", err)
	}

	return d.addMetadata(metadrive, func(machine *Machine) map[string]string {
		metadataMap := make(map[string]string)
		metadataMap["hostname"] = machine.Name
		for _, nic := range machine.Nics {
			if nic.Type == "veth" {
				metadataMap[fmt.Sprintf("%s_ipv4", nic.IfName)] = nic.Ipv4Addr
				metadataMap[fmt.Sprintf("%s_mac", nic.IfName)] = nic.MacAddr
			}
		}
		return metadataMap
	})
}

func (d *QEMUHypervisorDriver) CreateInstance() error {
	d.log().Infoln("Create instance...")
	instanceDir := filepath.Join(settings.InstancePath, d.Base.Instance.GetId())
	os.MkdirAll(instanceDir, os.ModePerm)

	instanceImage := d.machine.Drives["disk"].Image
	if _, err := os.Stat(instanceImage.Path) ; err != nil {
		d.log().Infoln("Create instance image...")
		baseImage, err := d.getImage()
		if  err != nil {
			return err
		}
		if err := instanceImage.SetBaseImage(baseImage); err != nil {
			return err
		}
		if err := instanceImage.CreateImage() ; err != nil {
			return err
		}
	}

	metadriveImage := d.machine.Drives["meta"].Image
	if _, err := os.Stat(metadriveImage.Path); err != nil {
		d.log().Infoln("Create metadrive image...")
		metadriveImage.SetSize(1440)
		if err := metadriveImage.CreateImage() ; err != nil {
			return err
		}
		if err := d.buildMetadriveBase(metadriveImage) ; err != nil {
			return err
			// todo remove metadrive image since it failed
		}
	}
	return nil
}

func (d *QEMUHypervisorDriver) DestroyInstance() error {
	if (d.machine.State == RUNNING) {
		d.StopInstance()
	}
	d.log().Infoln("Removing instance...")
	if err := os.RemoveAll(filepath.Join(settings.InstancePath, d.Base.Instance.GetId())) ; err != nil {
		return errors.Errorf("Failed to remove instance")
	}

	return nil
}

func (d *QEMUHypervisorDriver) StartInstance() error {
	d.log().Infoln("Starting qemu instance...")
	if err := d.machine.Start(filepath.Join(settings.QemuPath, settings.QemuProvider)); err != nil {
			return errors.Wrap(err, "Failed machine.Start()")
	}
	return nil
}

func (d *QEMUHypervisorDriver) StopInstance() error {
	d.log().Infoln("Stopping qemu instance...")
	if err := d.machine.Stop(); err != nil {
		return errors.Wrap(err, "Failed machine.Stop()")
	}
	return nil
}

func (d QEMUHypervisorDriver) RebootInstance() error {
	d.log().Infoln("Rebooting qemu instance...")
	if err := d.machine.Reboot(); err != nil {
		return errors.Wrap(err, "Failed machine.Reboot()")
	}
	return nil
}

func (d QEMUHypervisorDriver) InstanceConsole() hypervisor.Console {
	return nil
}
