package qemu

import (
	"github.com/pkg/errors"
	"strings"
)

// -device virtio-serial -chardev socket,path=/tmp/foo,server,nowait,id=foo -device virtserialport,chardev=foo,name=org.fedoraproject.port.0
type DeviceType int

const (
	DevType DeviceType = iota // 0
	NetType
	CharType
)

func (t DeviceType) String() string {
	switch t {
	case NetType:
		return "netdev"
	case CharType:
		return "chardev"
	default:
		return "device"
	}
}

type DriverOption struct {
	key   string
	value string
}

type Device struct {
	DeviceType string
	Params     *DeviceParams
}

type DeviceParams struct {
	Driver string
	// Wrap the key value pair here because maps are not sorted
	Options []DriverOption
}

func NewDevice(deviceType DeviceType) *Device {
	return &Device{
		DeviceType: deviceType.String(),
		Params: &DeviceParams{
			Options: make([]DriverOption, 0),
		},
	}
}

func (d *Device) AddDriver(driverType string) error {
	if len(d.Params.Driver) > 0 {
		return errors.Errorf("Driver has already been defined as %s", d.Params.Driver)
	}
	d.Params.Driver = driverType
	return nil
}

func (d *Device) AddDriverOption(key string, value string) error {
	for _, opt := range d.Params.Options {
		if opt.key == key {
			return errors.Errorf("key: %s is already assigned value %s", opt.key, opt.value)
		}
	}

	d.Params.Options = append(d.Params.Options, DriverOption{key: key, value: value})
	return nil
}

func (d *Device) LinkToGuestDevice(id string, guestDevice *Device) {
	d.AddDriverOption("id", id)
	guestDevice.AddDriverOption(d.DeviceType, id)
}

func (d *Device) BuildArg() []string {
	var arg []string
	var opt string
	arg = append(arg, strings.Join([]string{"-", d.DeviceType}, ""))
	if len(d.Params.Driver) > 0 {
		opt = d.Params.Driver
		for _, o := range d.Params.Options {
			opt = strings.Join([]string{opt, ",", o.key, "=", o.value}, "")
		}
	}
	arg = append(arg, opt)
	return arg
}
