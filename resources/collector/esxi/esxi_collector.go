package esxi

import (
	"fmt"
	"net/url"
	"bytes"
	"strconv"
	"io"
	"os"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
	"github.com/spf13/viper"
	"github.com/pkg/errors"
	cli "github.com/vmware/govmomi/govc/cli"
	_ "github.com/vmware/govmomi/govc/host"
	_ "github.com/vmware/govmomi/govc/host/esxcli"
)

type esxiResourceCollector struct {
	hostIp          string
	hostUser        string
	hostPass        string
	hostInsecure    bool
	datacenter      string
	esxiAPIEndpoint *url.URL
}

type esxiCmd func (cmd ...[]string) error

func init() {
	resources.RegisterCollector("esxi", NewEsxiResourceCollector)
}

// TODO: these functions were copy-pasted from the esxi driver and can
// probably be refactored into some util.
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

func (rm *esxiResourceCollector) esxiRunCmd(cmdList ...[]string) error {
	for _, args := range cmdList {
		a := []string{
			args[0],
			join('=', "-dc", rm.datacenter),
			join('=', "-k", strconv.FormatBool(rm.hostInsecure)),
			join('=', "-u", rm.esxiAPIEndpoint.String()),
		}
		for _, arg := range args[1:] {
			a = append(a, arg)
		}
		if rc := cli.Run(a); rc != 0 {
			return errors.Errorf("Failed api request: %s", args[0])
		}
	}
	return nil
}

func captureStdout(collector *esxiResourceCollector, cmd []string) ([]byte, error) {
	stdout := os.Stdout
	outputChan := make(chan []byte)

	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	os.Stdout = w
	restoreStdout := func() {
		w.Close()
		os.Stdout = stdout
	}


	go func() {
		b := make([]byte, 8192)
		var buf bytes.Buffer

		for {
			n, err := r.Read(b);
			if n == 0 {
				if err == io.EOF || err == nil {
					outputChan <-buf.Bytes()
					return
				}
			}
			buf.Write(b)
		}
	}()
	if err := collector.esxiRunCmd(cmd); err != nil {
		restoreStdout()
		return nil, err
	}

	restoreStdout()
	o := <-outputChan
	return o, nil
}

func NewEsxiResourceCollector(conf *viper.Viper) (resources.ResourceCollector, error) {
	initConfig(conf)
	var err error
	c := &esxiResourceCollector{
		hostIp:       conf.GetString("hypervisor.esxi-ip"),
		hostUser:     conf.GetString("hypervisor.esxi-user"),
		hostPass:     conf.GetString("hypervisor.esxi-pass"),
		hostInsecure: conf.GetBool("hypervisor.esxi-insecure"),
		datacenter:   conf.GetString("hypervisor.esxi-datacenter"),
	}

	uri := fmt.Sprintf("%s:%s@%s", c.hostUser, c.hostPass, c.hostIp)
	if c.esxiAPIEndpoint, err = url.Parse("https://" + uri + "/sdk"); err != nil {
		return nil, errors.Wrap(err, "Failed to parse url for ESXi server")
	}
	return c, nil
}

func (rm *esxiResourceCollector) GetCpu() (*model.Resource, error) {
	// get in use:
	// ./govc host.info -host vmware -json | jq -r .HostSystems[].Summary.QuickStats.OverallCpuUsage
	cmd := []string{"host.esxcli", "-json", "hardware", "cpu", "global", "get"}
	output, err := captureStdout(rm, cmd)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output))
	return &model.Resource{}, nil
}

func (rm *esxiResourceCollector) GetMem() (*model.Resource, error) {
	// get total ./govc host.info -host vmware -json | jq -r .HostSystems[].Summary.Hardware.MemorySize
	// get in use ./govc host.info -host vmware -json | jq -r .HostSystems[].Summary.QuickStats.OverallMemoryUsage
	// available = total - used
	// convert used into bytes
	// percent =  ((used << 2e) * 100
	cmd := []string{"host.info", "-json"}
	output, err := captureStdout(rm, cmd)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output))
	return &model.Resource{}, nil
}

func (rm *esxiResourceCollector) GetDisk() ([]*model.Resource, error) {
	// ./govc host.esxcli -json storage filesystem list | jq .Values
	cmd := []string{"host.esxcli", "-json", "storage", "filesystem", "list"}
	output, err := captureStdout(rm, cmd)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output))
	disks := make([]*model.Resource, 0)
	return disks, nil
}

func (rm *esxiResourceCollector) GetLoadAvg() (*model.LoadAvg, error) {
	// ./govc host.esxcli -json system process stats load get | jq -r .Values

	cmd := []string{"host.esxcli", "-json", "system", "process", "stats", "load", "get"}
	output, err := captureStdout(rm, cmd)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output))
	return &model.LoadAvg{}, nil
}
