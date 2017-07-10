package qemu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDevice(t *testing.T) {
	assert := assert.New(t)
	device := NewDevice(DevType) 
	assert.NotNil(device)
	assert.Equal("device", device.DeviceType)
	assert.NotNil(device.Params)
}

func TestAddDriver(t *testing.T) {
	assert := assert.New(t)
	device := NewDevice(DevType)
	device.AddDriver("tap")
	assert.NotNil(device.Params.Driver)
	assert.Equal(device.Params.Driver, "tap")
}

func TestAddDriverOption(t *testing.T) {
	assert := assert.New(t)
	device := NewDevice(DevType)
	device.AddDriverOption("key", "value")
	assert.NotNil(device.Params.Options[0])
	assert.Equal(len(device.Params.Options), 1)
	assert.Equal("key", device.Params.Options[0].key)
	assert.Equal("value", device.Params.Options[0].value)
}

func TestLinkToGuestDevice(t *testing.T) {
	assert := assert.New(t)
	device1 := NewDevice(NetType)
	device2 := NewDevice(DevType)
	device1.LinkToGuestDevice("dev", device2)

	assert.Equal(device1.Params.Options[0].key, "id")
	assert.Equal(device1.Params.Options[0].value, "dev")

	assert.Equal(device2.Params.Options[0].key, "netdev")
	assert.Equal(device2.Params.Options[0].value, "dev")
}

func TestBuildArg(t *testing.T) {
	assert := assert.New(t)
	device := NewDevice(DevType)
	device.AddDriver("driver")
	device.AddDriverOption("id", "driver")
	args := device.EvaluateCliCmd()

	assert.NotNil(args)
	assert.Equal(len(args), 2)
}
