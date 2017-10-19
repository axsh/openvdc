package esxi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"bytes"
	"strconv"
	"io"
	"reflect"
	"os"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
	"github.com/spf13/viper"
	"github.com/pkg/errors"
	"github.com/Jeffail/gabs"
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
	datastore       string

	esxiAPIEndpoint *url.URL
}

type esxiCmd func (cmd ...[]string) error

func init() {
	resources.RegisterCollector("esxi", NewEsxiResourceCollector)
}

func initConfig(conf *viper.Viper) error {
	viper.SetDefault("hypervisor.esxi-insecure", true)

	if conf.GetString("hypervisor.esxi-user") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-user")
	}
	if conf.GetString("hypervisor.esxi-pass") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-pass")
	}
	if conf.GetString("hypervisor.esxi-ip") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-ip")
	}
	if conf.GetString("hypervisor.esxi-datacenter") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-datacenter")
	}
	if conf.GetString("hypervisor.esxi-datastore") == "" {
		return errors.Errorf("Missing configuration hypervisor.exsi-datastore")
	}
	return nil
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
			} else {
				outputChan <- func() ([]byte, error) {
					return nil, errors.Wrap(err, "Failed io.Copy()")
				}
			}
			return
		}
	}()

	if err := fn(); err != nil {
		return nil, err
	}
	w.Close()
	os.Stdout = stdout
	return (<-outputChan)()
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

func (rm *esxiResourceCollector) esxiRunCmd(args []string) ([]byte, error) {
	return captureStdout(func() error {
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
		return nil
	})
}

// workaround to reach the actual 'interface{}' because the esxi
// returns everything as slices, and sometimes nested slices
func typeAssert(object *gabs.Container) interface{} {
	newObject := object.Data().([]interface{})[0]
	for {
		t := reflect.TypeOf(newObject)
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			newObject = newObject.([]interface{})[0]
		default:
			return newObject
		}
	}
	return nil
}

func NewEsxiResourceCollector(conf *viper.Viper) (resources.ResourceCollector, error) {
	if err := initConfig(conf); err != nil {
		return nil, errors.Wrap(err, "Failed to initialize config")
	}
	var err error
	c := &esxiResourceCollector{
		hostIp:       conf.GetString("hypervisor.esxi-ip"),
		hostUser:     conf.GetString("hypervisor.esxi-user"),
		hostPass:     conf.GetString("hypervisor.esxi-pass"),
		hostInsecure: conf.GetBool("hypervisor.esxi-insecure"),
		datacenter:   conf.GetString("hypervisor.esxi-datacenter"),
		datastore:    conf.GetString("hypervisor.esxi-datastore"),
	}

	uri := fmt.Sprintf("%s:%s@%s", c.hostUser, c.hostPass, c.hostIp)
	if c.esxiAPIEndpoint, err = url.Parse("https://" + uri + "/sdk"); err != nil {
		return nil, errors.Wrap(err, "Failed to parse url for ESXi server")
	}
	return c, nil
}

func (rm *esxiResourceCollector) GetCpu() (*model.Resource, error) {
	cmd := []string{"host.esxcli", "-json", "hardware", "cpu", "global", "get"}
	output, err := rm.esxiRunCmd(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "Faield to capture stdout")
	}

	parsedJson, err := gabs.ParseJSON(output)
	if err != nil {
		return nil, err
	}
	v := parsedJson.Path("Values")
	cores, _ := strconv.ParseInt(typeAssert(v.Path("CPUCores")).(string), 10, 64)
	// pkgs, _ := strconv.ParseInt(typeAssert(v.Path("CPUPackages")).(string),10, 64)
	return &model.Resource{
		Total: cores,
	}, nil
}

func (rm *esxiResourceCollector) GetMem() (*model.Resource, error) {
	cmd := []string{"host.info", "-json"}
	output, err := rm.esxiRunCmd(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "Faield to capture stdout")
	}

	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.UseNumber()
	parsedJson, err := gabs.ParseJSONDecoder(decoder)
	if err != nil {
		return nil, err
	}
	s := parsedJson.Path("HostSystems.Summary")
	total, _ := typeAssert(s.Path("Hardware.MemorySize")).(json.Number).Int64()
	used, _ := typeAssert(s.Path("QuickStats.OverallMemoryUsage")).(json.Number).Int64()

	return &model.Resource{
		Total: total,
		Available: (total - (used << 20)),
		UsedPercent: ((float64(used << 20) / float64(total)) * 100),
	}, nil
}

func (rm *esxiResourceCollector) GetDisk() ([]*model.Resource, error) {
	cmd := []string{"host.esxcli", "-json", "storage", "filesystem", "list"}
	output, err := rm.esxiRunCmd(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "Faield to capture stdout")
	}

	disks := make([]*model.Resource, 0)
	parsedJson, err := gabs.ParseJSON(output)
	if err != nil {
		return nil, err
	}
	v, _ := parsedJson.Path("Values").Children()
	for _, child := range v {
		storageDisk := typeAssert(child.Path("VolumeName")).(string)
		if storageDisk != rm.datastore {
			continue
		}
		free, _ := strconv.ParseInt(typeAssert(child.Path("Free")).(string), 10, 64)
		size, _ := strconv.ParseInt(typeAssert(child.Path("Size")).(string), 10, 64)
		disks = append(disks, &model.Resource{
			Total: size,
			Available: free,
			UsedPercent: ((float64(size - free) / float64(size)) * 100),
		})
	}

	return disks, nil
}

func (rm *esxiResourceCollector) GetLoadAvg() (*model.LoadAvg, error) {
	cmd := []string{"host.esxcli", "-json", "system", "process", "stats", "load", "get"}
	output, err := rm.esxiRunCmd(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "Faield to capture stdout")
	}

	parsedJson, err := gabs.ParseJSON(output)
	if err != nil {
		return nil, err
	}
	v := parsedJson.Path("Values")
	load1, _ := strconv.ParseFloat(typeAssert(v.Path("Load1Minute")).(string), 32)
	load5, _ := strconv.ParseFloat(typeAssert(v.Path("Load5Minutes")).(string), 32)
	load15, _ := strconv.ParseFloat(typeAssert(v.Path("Load15Minutes")).(string), 32)

	return &model.LoadAvg{
		Load1:  float32(load1),
		Load5:  float32(load5),
		Load15: float32(load15),
	}, nil
}
