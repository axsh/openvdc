package esxi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/hypervisor/util"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	cli "github.com/vmware/govmomi/govc/cli"
	_ "github.com/vmware/govmomi/govc/datastore"
	_ "github.com/vmware/govmomi/govc/device"
	_ "github.com/vmware/govmomi/govc/device/floppy"
	_ "github.com/vmware/govmomi/govc/device/serial"
	_ "github.com/vmware/govmomi/govc/vm"
	_ "github.com/vmware/govmomi/govc/vm/guest"
	_ "github.com/vmware/govmomi/govc/vm/network"
	"golang.org/x/crypto/ssh"
)

type BridgeType int

const (
	None BridgeType = iota // 0
	Linux
	OVS
)

var settings struct {
	ScriptPath          string
	EsxiUser            string
	EsxiPass            string
	EsxiIp              string
	EsxiDatacenter      string
	EsxiHostName        string
	EsxiInsecure        bool
	EsxiHostSshkey      string
	EsxiVmUser          string
	EsxiVmPass          string
	EsxiVmDatastore     string
	EsxiUrl             string
	EsxiInventoryFolder string
	vCenterEndpoint     bool
	BridgeName          string
	BridgeType          BridgeType
}

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

type EsxiHypervisorProvider struct {
}

type EsxiHypervisorDriver struct {
	hypervisor.Base
	template  *model.EsxiTemplate
	machine   *EsxiMachine
	imageName string
	hostName  string
	vmName    string
}

func (p *EsxiHypervisorProvider) Name() string {
	return "esxi"
}

func init() {
	hypervisor.RegisterProvider("esxi", &EsxiHypervisorProvider{})
	viper.SetDefault("hypervisor.esxi-insecure", true)
	viper.SetDefault("hypervisor.vCenter", false)
}

func (p *EsxiHypervisorProvider) LoadConfig(sub *viper.Viper) error {
	if sub.IsSet("bridges.name") {
		settings.BridgeName = sub.GetString("bridges.name")
		if sub.IsSet("bridges.type") {
			switch sub.GetString("bridges.type") {
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
	} else if sub.IsSet("bridges.ovs.name") {
		log.Warn("bridges.ovs.name is obsolete option")
		settings.BridgeName = sub.GetString("bridges.ovs.name")
		settings.BridgeType = OVS
	}
	if sub.GetString("hypervisor.esxi-datacenter") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-datacenter")
	}
	settings.EsxiDatacenter = sub.GetString("hypervisor.esxi-datacenter")

	if sub.GetString("hypervisor.esxi-user") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-user")
	}
	settings.EsxiUser = sub.GetString("hypervisor.esxi-user")

	if sub.GetString("hypervisor.esxi-pass") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-pass")
	}
	settings.EsxiPass = sub.GetString("hypervisor.esxi-pass")

	if sub.GetString("hypervisor.esxi-ip") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-ip")
	}
	settings.EsxiIp = sub.GetString("hypervisor.esxi-ip")

	if sub.GetString("hypervisor.esxi-vm-datastore") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-vm-datastore")
	}
	settings.EsxiVmDatastore = sub.GetString("hypervisor.esxi-vm-datastore")

	settings.vCenterEndpoint = sub.GetBool("hypervisor.vCenter")
	settings.ScriptPath = sub.GetString("hypervisor.script-path")
	settings.EsxiInsecure = sub.GetBool("hypervisor.esxi-insecure")
	settings.EsxiVmUser = sub.GetString("hypervisor.esxi-vm-user")
	settings.EsxiVmPass = sub.GetString("hypervisor.esxi-vm-pass")
	settings.EsxiInventoryFolder = sub.GetString("hypervisor.esxi-inventory-folder")
	if settings.vCenterEndpoint {
		if sub.GetString("hypervisor.esxi-host-name") == "" {
			return errors.Errorf("Missing configuration hypervisor.esxi-host-name")
		}
		settings.EsxiHostName = sub.GetString("hypervisor.esxi-host-name")
	} else {
		if sub.GetString("hypervisor.esxi-host-sshkey") == "" {
			return errors.Errorf("Missing configuration hypervisor.esxi-host-sshkey")
		}
		settings.EsxiHostSshkey = sub.GetString("hypervisor.esxi-host-sshkey")

	}

	esxiInfo := fmt.Sprintf("%s:%s@%s", settings.EsxiUser, settings.EsxiPass, settings.EsxiIp)
	u, err := url.Parse("https://" + esxiInfo + "/sdk")
	if err != nil {
		return errors.Wrap(err, "Failed to parse url for ESXi server")
	}
	settings.EsxiUrl = u.String()

	return nil
}

func (p *EsxiHypervisorProvider) CreateDriver(instance *model.Instance, template model.ResourceTemplate) (hypervisor.HypervisorDriver, error) {
	esxiTmpl, ok := template.(*model.EsxiTemplate)
	if !ok {
		return nil, errors.Errorf("template type is not *model.WmwareTemplate: %T, template")
	}
	instanceIdx, _ := strconv.Atoi(strings.TrimPrefix(instance.GetId(), "i-"))
	driver := &EsxiHypervisorDriver{
		Base: hypervisor.Base{
			Log:      log.WithFields(log.Fields{"Hypervisor": "esxi", "instance_id": instance.GetId()}),
			Instance: instance,
		},
		template: esxiTmpl,
		vmName:   instance.GetId(),
	}
	driver.machine = newEsxiMachine(15000+instanceIdx, driver.template)
	return driver, nil
}

func (d *EsxiHypervisorDriver) log() *log.Entry {
	return d.Base.Log
}

func join(separator byte, args ...string) string {
	argLength := len(args)
	currentArg := 0
	var buf bytes.Buffer
	for _, arg := range args {
		currentArg = currentArg + 1
		buf.WriteString(arg)
		if currentArg == argLength {
			separator = 0
		}
		if separator > 0 {
			buf.WriteByte(separator)
		}
	}
	return buf.String()
}

func captureStdout(fn func() error) ([]byte, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, errors.Wrap(err, "Failed os.Pipe()")
	}
	stdout := os.Stdout
	os.Stdout = w

	outputChan := make(chan func() ([]byte, error))
	go func() {
		var buf bytes.Buffer
		if n, err := io.Copy(&buf, r); n > 0 {
			if err == nil || err == io.EOF {
				outputChan <- func() ([]byte, error) {
					return buf.Bytes(), nil
				}
				return
			} else {
				outputChan <- func() ([]byte, error) {
					return nil, errors.Wrap(err, "Failed io.Copy()")
				}
				return
			}
		}
		outputChan <- func() ([]byte, error) { return nil, nil }
	}()

	if err := fn(); err != nil {
		return nil, err
	}
	w.Close()
	os.Stdout = stdout
	return (<-outputChan)()
}

func runCmd(cmd string, args []string) error {
	c := exec.Command(cmd, args...)
	if err := c.Run(); err != nil {
		return errors.Errorf("failed to execute command :%s %s", cmd, args)
	}
	return nil
}

var ErrApiRequest = errors.New("Failed api requiest")
func esxiRunCmd(cmdList ...[]string) error {
	for _, args := range cmdList {
		a := []string{
			args[0],
			join('=', "-dc", settings.EsxiDatacenter),
			join('=', "-k", strconv.FormatBool(settings.EsxiInsecure)),
			join('=', "-u", settings.EsxiUrl),
		}
		for _, arg := range args[1:] {
			a = append(a, arg)
		}
		log.Info("Calling:", a)
		if rc := cli.Run(a); rc != 0 {
			log.Errorf("failed request: %s", args[0])
			return ErrApiRequest
		}
	}
	return nil
}

func deviceExists(vm string, device string) (bool, error) {
	exists := false

	// get in json format, normal stdout is unreliable
	output, err := captureStdout(func() error {
		return esxiRunCmd([]string{"device.info", "-json", vm, device})
	})
	if err != nil {
		// device.info can also fail when it cannot find any matching device
		if err == ErrApiRequest {
			return exists, nil
		} else {
			return exists, errors.Errorf("failed captureStdout()", err)
		}
	}

	var dev struct {
		Devices []interface{} `json="Devices,omitempty"`
	}
	if err := json.Unmarshal(output, &dev); err != nil {
		return exists, errors.Errorf("Failed json.Unmarshal", err)
	}
	if dev.Devices != nil {
		exists = true
	}
	return exists, nil
}

// wrappers for esxi api syntax
func storageImg(imgName string) string {
	path := join('/', imgName, imgName)
	return join('.', path, "vmx")
}

func vmUserDetails() string {
	userDetails := join(':', settings.EsxiVmUser, settings.EsxiVmPass)
	return join('=', "-l", userDetails)
}

func (d *EsxiHypervisorDriver) vmPath() string {
	return join(0, "-vm.path=[", settings.EsxiVmDatastore, "]", storageImg(d.vmName))
}

func (d *EsxiHypervisorDriver) MetadataDrivePath() string {
	return "/tmp/metadrive.img"
}

func (d *EsxiHypervisorDriver) MetadataDriveDatamap() map[string]interface{} {
	metadataMap := make(map[string]interface{})
	metadataMap["hostname"] = d.vmName
	for idx, nic := range d.machine.Nics {
		iface := make(map[string]interface{})
		iface["ifname"] = nic.IfName
		iface["ipv4"] = nic.Ipv4Addr
		iface["mac"] = nic.MacAddr
		metadataMap[fmt.Sprintf("nic-%02d", idx)] = iface
	}
	return metadataMap
}

func (d *EsxiHypervisorDriver) CreateInstance() error {
	var err error
	for idx, iface := range d.template.GetInterfaces() {
		d.machine.Nics = append(d.machine.Nics, Nic{
			NetworkId: iface.NetworkId,
			IfName:    fmt.Sprintf("%s_%02d", d.vmName, idx),
			Type:      iface.Type,
			Ipv4Addr:  iface.Ipv4Addr,
			MacAddr:   iface.Macaddr,
			Bridge:    settings.BridgeName,
		})
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
	if err = d.CloneBaseImage(); err != nil {
		return err
	}
	err = esxiRunCmd(
		[]string{"datastore.upload", join('=', "-ds", settings.EsxiVmDatastore), d.MetadataDrivePath(), fmt.Sprintf("%s/metadrive.img", d.vmName)},
	)
	if err != nil {
		return err
	}
	if err := os.Remove(d.MetadataDrivePath()); err != nil {
		return errors.Errorf("Unable to remove metadrive: %s", d.MetadataDrivePath())
	}

	// NOTE:
	// serial port devices starts from 9000
	// floppy devices starts from 8000
	if err = d.AddFloppyDevice(); err != nil {
		return err
	}
	if err = d.AddNetworkDevices(); err != nil {
		return err
	}
	err = esxiRunCmd(
		[]string{"device.floppy.insert", "-device=floppy-8000", fmt.Sprintf("-vm=%s", d.vmName), fmt.Sprintf("%s/metadrive.img", d.vmName)},
		[]string{"device.serial.add", d.vmPath()},
		[]string{"device.serial.connect", d.vmPath(), "-device=serialport-9000", join(':', "telnet://", strconv.Itoa(d.machine.SerialConsolePort))},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *EsxiHypervisorDriver) CloneBaseImage() error {
	datastore := join('=', "-ds", settings.EsxiVmDatastore)

	if settings.vCenterEndpoint {
		cmd := []string{"vm.clone", datastore,
			join('=', "-on", "false"),
			join('=', "-host", settings.EsxiHostName),
			join('=', "-vm", d.machine.baseImage.name),
		}
		if len(settings.EsxiInventoryFolder) > 0 {
			cmd = append(cmd, join('=', "-folder", settings.EsxiInventoryFolder))
		}
		cmd = append(cmd, d.vmName)
		err := esxiRunCmd(cmd)
		if err != nil {
			return err
		}
		return nil
	} else {
		if d.machine.baseImage.datastore == "" {
			return errors.Errorf("Empty path to datastore for vm: %s", d.machine.baseImage.name)
		}
		err := esxiRunCmd(
			[]string{"datastore.mkdir", join('=', "-ds", settings.EsxiVmDatastore), d.vmName},
		)
		if err != nil {
			return err
		}

		key, err := ioutil.ReadFile(settings.EsxiHostSshkey)
		if err != nil {
			return errors.Errorf("Unable to read the specified ssh private key: %s", settings.EsxiHostSshkey)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return errors.Errorf("Unable to parse the specified ssh private key: %s", settings.EsxiHostSshkey)
		}

		conn, err := ssh.Dial("tcp", strings.Join([]string{settings.EsxiIp, "22"}, ":"), &ssh.ClientConfig{
			User: settings.EsxiUser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		})

		if err != nil {
			return errors.Errorf("Unable establish ssh connection to %s", settings.EsxiIp)
		}
		defer conn.Close()

		session, err := conn.NewSession()
		if err != nil {
			return errors.Errorf("Unable to open a session on connection %d", conn)
		}
		defer session.Close()

		var out bytes.Buffer
		var stderr bytes.Buffer
		session.Stdout = &out
		session.Stderr = &stderr

		// Ssh into esxiHost and use "vmkfstools" to clone vmdk"
		basePath := join('/', "/vmfs", "volumes", d.machine.baseImage.datastore, d.machine.baseImage.name, strings.Join([]string{d.machine.baseImage.name, ".vmdk"}, ""))
		newPath := join('/', "/vmfs", "volumes", settings.EsxiVmDatastore, d.vmName, strings.Join([]string{d.machine.baseImage.name, ".vmdk"}, ""))

		if err := session.Run(join(' ', "vmkfstools -i", basePath, newPath, "-d thin")); err != nil {
			return errors.Errorf(stderr.String(), "Error cloning vmdk")
		}
		err = esxiRunCmd(
			[]string{"datastore.cp", datastore, storageImg(d.machine.baseImage.name), storageImg(d.vmName)},
			[]string{"vm.register", datastore, storageImg(d.vmName)},
			[]string{"vm.change", join('=', "-name", d.vmName), d.vmPath()},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *EsxiHypervisorDriver) AddNetworkDevices() error {
	for _, nic := range d.machine.Nics {
		cmd := []string{"vm.network.add", d.vmPath(), join('=', "-net.adapter", nic.Type)}

		if len(nic.NetworkId) > 0 {
			networkId := join('=', "-net", nic.NetworkId)
			exists, err := deviceExists(d.vmPath(), networkId)
			if err != nil {
				return err
			}
			if exists {
				log.Infof("Machine already has an adapter in network %s attached, skipping", nic.NetworkId)
				continue
			}

			cmd = append(cmd, networkId)
		}
		if len(nic.MacAddr) > 0 {
			cmd = append(cmd, join('=', "-net.address", nic.MacAddr))
		}
		if err := esxiRunCmd(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (d *EsxiHypervisorDriver) AddFloppyDevice() error {
	exists, err := deviceExists(d.vmPath(), "floppy-*")
	if err != nil {
		return err
	}
	if exists {
		log.Infof("Machine already has floppy device, skipping")
		return nil
	}
	return esxiRunCmd([]string{"device.floppy.add", d.vmPath()})
}

func (d *EsxiHypervisorDriver) DestroyInstance() error {
	return esxiRunCmd(
		[]string{"datastore.rm", fmt.Sprintf("-ds=%s", settings.EsxiVmDatastore), fmt.Sprintf("%s/metadrive.img", d.vmName)},
		[]string{"vm.destroy", d.vmPath()},
	)
}

func (d *EsxiHypervisorDriver) StartInstance() error {
	port := strconv.Itoa(d.machine.SerialConsolePort)

	return esxiRunCmd(
		[]string{"device.serial.connect", d.vmPath(), "-device=serialport-9000", join(':', "telnet://", port)},
		[]string{"vm.power", "-on=true", "-suspend=false", d.vmPath()},
	)
}

func (d *EsxiHypervisorDriver) StopInstance() error {
	return esxiRunCmd(
		[]string{"vm.power", "-suspend=true", d.vmPath()},
		[]string{"device.serial.disconnect", d.vmPath(), fmt.Sprintf("-device=serialport-9000")},
	)
}

func (d EsxiHypervisorDriver) RebootInstance() error {
	// Linux, this should be doable through api call.
	return esxiRunCmd(
		[]string{"guest.start", vmUserDetails(), d.vmPath(), "/sbin/reboot"},
	)
}

func (d *EsxiHypervisorDriver) Recover(instanceState model.InstanceState) error {
	//Todo: handle recovery
	return nil
}
