package qemu

import (
	"strings"
	"github.com/pkg/errors"
)

// -device virtio-serial -chardev socket,path=/tmp/foo,server,nowait,id=foo -device virtserialport,chardev=foo,name=org.fedoraproject.port.0

type DriverOption struct {
	key   string
	value string
}

type Device struct {
	Type        string
	Id          string
	Params      *DeviceParams
	GuestDevice *Device
}

type DeviceParams struct {
	Driver   string

	// Wrap the key value pair here because maps are not sorted
	Options  []DriverOption
}

func NewDevice(id string) *Device {
	return &Device{
		Type: "device",
		Id: id,
		Params: &DeviceParams{
			Options: make([]DriverOption, 0),
		},
	}
}


func (d *Device) SetDeviceType(deviceType string) {
	d.Type = deviceType
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

func (d *Device) AddGuestDevice(deviceType string) {
	d.GuestDevice = NewDevice(d.Id)
	d.GuestDevice.AddDriver(deviceType)
	d.GuestDevice.AddDriverOption(d.Type, d.GuestDevice.Id)
}

func (d *Device) EvaluateCliCmd() string {
	var opt string
	arg := strings.Join([]string{"-", d.Type}, "")
	if len(d.Params.Driver) > 0 {
		opt = d.Params.Driver
		for _, o := range d.Params.Options {
			opt = strings.Join([]string{opt, ",", o.key, "=", o.value}, "")
		}
	}
	if d.GuestDevice != nil {
		opt = strings.Join([]string{opt, d.GuestDevice.EvaluateCliCmd()}, " ")
	}
	return strings.Join([]string{arg, opt}, " ")
}
