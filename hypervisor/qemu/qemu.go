// +build linux

package qemu

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/hypervisor/util"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
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
	machine   *Machine
}

func (p *QEMUHypervisorProvider) Name() string {
	return "qemu"
}

var settings struct {
	ImageServerUri   string
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
	if _, err := os.Stat("/usr/libexec/qemu-kvm"); err == nil {
		settings.QemuPath = "/usr/libexec"
		settings.QemuProvider = "qemu-kvm"
	} else if _, err := os.Stat("/usr/bin/qemu-system-x86_64"); err == nil {
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

func (p *QEMUHypervisorProvider) CreateDriver(instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	qemuTmpl, ok := template.(*model.QemuTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.QemuTemplate: %T", template)
	}
	driver := &QEMUHypervisorDriver{
		Base: hypervisor.Base{
			Log:      log.WithFields(log.Fields{"Hypervisor": "qemu", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: qemuTmpl,
	}
	driver.createMachineTemplate()
	return driver, nil
}

func (d *QEMUHypervisorDriver) createMachineTemplate() {
	instanceId := d.Base.Instance.GetId()
	instanceDir := filepath.Join(settings.InstancePath, instanceId)
	imageFormat := strings.ToLower(d.template.QemuImage.GetFormat().String())
	var netDev []Nic
	for idx, iface := range d.template.GetInterfaces() {
		netDev = append(netDev, Nic{
			IfName:       fmt.Sprintf("%s_%02d", instanceId, idx),
			Type:         iface.Type,
			Ipv4Addr:     iface.Ipv4Addr,
			MacAddr:      iface.Macaddr,
			Bridge:       settings.BridgeName,
			BridgeHelper: settings.QemuBridgeHelper,
		})
	}

	d.machine = NewMachine(int(d.template.GetVcpu()), uint64(d.template.GetMemoryGb()*1024))
	d.machine.Drives["disk"] = Drive{Image: NewImage(filepath.Join(instanceDir, "diskImage."+imageFormat), imageFormat)}
	d.machine.Drives["meta"] = Drive{Image: NewImage(filepath.Join(instanceDir, "metadrive.img"), "raw"), If: "floppy"}
	d.machine.Name = instanceId
	d.machine.MonitorSocketPath = filepath.Join(instanceDir, "monitor.socket")
	d.machine.SerialSocketPath = filepath.Join(instanceDir, "serial.socket")
	d.machine.Kvm = d.template.GetUseKvm()
	for _, dev := range d.machine.AddNICs(netDev) {
		d.machine.AddDevice(dev)
	}

	// these devices are required for communication through the qemu guest agent
	d.machine.AgentSocketPath = filepath.Join(instanceDir, "agent.socket")
	virtioserialDev := NewDevice(DevType)
	virtioserialDev.AddDriver("virtio-serial")

	hostDev := NewDevice(CharType)
	hostDev.AddDriver("socket")
	hostDev.AddDriverOption("path", fmt.Sprintf("%s,server,nowait", d.machine.AgentSocketPath))

	guestDev := NewDevice(DevType)
	guestDev.AddDriver("virtserialport")
	guestDev.AddDriverOption("name", "org.qemu.guest_agent.0")

	hostDev.LinkToGuestDevice(instanceId, guestDev)

	d.machine.AddDevice(virtioserialDev)
	d.machine.AddDevice(hostDev)
	d.machine.AddDevice(guestDev)
}

func (d *QEMUHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func url2Local(downloadUrl string) (string, error) {
	parsedUrl, _ := url.Parse(downloadUrl)
	localPath := settings.CachePath

	if govalidator.IsRequestURL(downloadUrl) {
		localPath = filepath.Join(localPath, parsedUrl.Host, parsedUrl.Path)
		if _, err := os.Stat(localPath); err != nil {
			// strip the filename from path
			tmp := strings.Split(localPath, "/")
			if err := os.MkdirAll(strings.Join(tmp[0:len(tmp)-1], "/"), os.ModePerm); err != nil {
				return "", errors.Errorf("Unable to create folder: %s", strings.Join(tmp[0:len(tmp)-1], "/"))
			}
		}
	} else {
		return "", errors.Errorf("Unable to resolve download_url: %s", parsedUrl.String())
	}
	return localPath, nil
}

func (d *QEMUHypervisorDriver) getImage() (string, error) {
	imageCachePath, err := url2Local(d.template.QemuImage.GetDownloadUrl())
	remotePath := d.template.QemuImage.GetDownloadUrl()
	if err != nil {
		if settings.ImageServerUri != "" {
			imageCachePath = filepath.Join(settings.CachePath, remotePath)
			remotePath = settings.ImageServerUri + "/" + remotePath
		} else {
			return "", err
		}
	}

	if _, err := os.Stat(imageCachePath); err != nil {
		d.log().Infoln("Downloading machine image...")

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

func (d *QEMUHypervisorDriver) MetadataDrivePath() string {
	return d.machine.Drives["meta"].Image.Path
}

func (d *QEMUHypervisorDriver) MetadataDriveDatamap() map[string]interface{}{
	metadataMap := make(map[string]interface{})
	metadataMap["hostname"] = d.machine.Name
	for idx, nic := range d.machine.Nics {
		if nic.Type == "veth" {
			iface := make(map[string]interface{})
			iface["ifname"] = nic.IfName
			iface["ipv4"] = nic.Ipv4Addr
			iface["gateway"] = nic.Gateway
			iface["mac"] = nic.MacAddr
			metadataMap[fmt.Sprintf("nic-%02d", idx)] = iface
		}
	}
	return metadataMap
}

func (d *QEMUHypervisorDriver) Recover(instanceState model.InstanceState) error {
	//Todo: handle recovery
	return nil
}

func (d *QEMUHypervisorDriver) CreateInstance() error {
	d.log().Infoln("Create instance...")
	instanceDir := filepath.Join(settings.InstancePath, d.Base.Instance.GetId())

	if err := os.MkdirAll(instanceDir, os.ModePerm); err != nil {
		return errors.Errorf("Unable to create folder: %s", instanceDir)
	}
	if _, ok := d.machine.Drives["disk"]; !ok {
		return errors.Errorf("Faild to assign disk image")
	}
	instanceImage := d.machine.Drives["disk"].Image
	if _, err := os.Stat(instanceImage.Path); err != nil {
		d.log().Infoln("Create instance image...")
		baseImage, err := d.getImage()
		if err != nil {
			return err
		}
		if err := instanceImage.SetBaseImage(baseImage); err != nil {
			return err
		}
		if err := instanceImage.CreateImage(); err != nil {
			return err
		}
	}

	if _, ok := d.machine.Drives["meta"]; !ok {
		return errors.Errorf("Failed to assing metadrive image")
	}
	metadriveImage := d.machine.Drives["meta"].Image
	if _, err := os.Stat(d.MetadataDrivePath()); err != nil {
		d.log().Infoln("Create metadrive image...")
		metadriveImage.SetSize(1440)
		if err := metadriveImage.CreateImage(); err != nil {
			return err
		}
		if err := util.CreateMetadataDisk(d); err != nil {
			return err
		}
		if err := util.MountMetadataDisk(d); err != nil {
			return err
		}
		if err := util.WriteMetadata(d); err != nil {
			return err
		}
		if err := util.UmountMetadataDisk(d); err != nil {
			return err
		}
	}
	return nil
}

func (d *QEMUHypervisorDriver) DestroyInstance() error {
	// For now HavePrompt == RUNNING
	if d.machine.HavePrompt() {
		d.StopInstance()
	}
	d.log().Infoln("Removing instance...")
	if err := os.RemoveAll(filepath.Join(settings.InstancePath, d.Base.Instance.GetId())); err != nil {
		return errors.Errorf("Failed to remove instance")
	}

	return nil
}

func (d *QEMUHypervisorDriver) StartInstance() error {
	d.log().Infoln("Starting qemu instance...")
	if err := d.machine.Start(filepath.Join(settings.QemuPath, settings.QemuProvider)); err != nil {
		return errors.Wrap(err, "Failed machine.Start()")
	}

	return d.machine.ScheduleState(RUNNING, 10*time.Minute, func() bool {
		return d.machine.WaitForPrompt()
	})
}

func (d *QEMUHypervisorDriver) StopInstance() error {
	d.log().Infoln("Stopping qemu instance...")
	if err := d.machine.Stop(); err != nil {
		return errors.Wrap(err, "Failed machine.Stop()")
	}

	return d.machine.ScheduleState(STOPPED, 10*time.Minute, func() bool {
		// TODO: if states are to be kept, implement a way to evaluate that the instance is stopped (check process id? check serial i/o?)
		return true
	})
}

func (d QEMUHypervisorDriver) RebootInstance() error {
	d.log().Infoln("Rebooting qemu instance...")
	if err := d.machine.Reboot(); err != nil {
		return errors.Wrap(err, "Failed machine.Reboot()")
	}
	return d.machine.ScheduleState(RUNNING, 10*time.Minute, func() bool {
		return d.machine.WaitForPrompt()
	})
}

