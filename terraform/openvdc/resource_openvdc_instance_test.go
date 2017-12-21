// +build terraform

package openvdc 

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var getInterfaces = func() interface{} {
	return []interface{}{
		map[string]interface{}{
			"type": "veth",
			"ipv4addr": "0.0.0.0",
		},
		map[string]interface{}{
			"type": "veth",
			"ipv4addr": "1.0.0.0",
		},
	}
}
var getResources = func() interface{} {
	return map[string]interface{}{
		"vcpu": 1,
		"memory_gb": 1024,
	}
}
func TestRenderInterfaceOpt(t *testing.T) {
	assert := assert.New(t)

	parsedInterfaces, err := renderInterfaceOpt(getInterfaces)
	assert.NoError(err)
	assert.Equal(string(parsedInterfaces), `"interfaces":[{"ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]`)
}

func TestRenderResourceOpt(t *testing.T) {
	assert := assert.New(t)

	parsedResources, err := renderResourceOpt(getResources)
	assert.NoError(err)
	assert.Equal(string(parsedResources), `"memory_gb":1024,"vcpu":1`)
}

func TestRenderCmdOpt(t *testing.T) {
	assert := assert.New(t)

	p := []option{
		func() (renderCallback, resourceCallback) { return renderResourceOpt, getResources },
		func() (renderCallback, resourceCallback) { return renderInterfaceOpt, getInterfaces },
	}
	cmdOpts, err := renderCmdOpt(p)
	assert.NoError(err)
	assert.Equal(cmdOpts.String(), `{"memory_gb":1024,"vcpu":1,"interfaces":[{"ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]}`)
	fmt.Println(cmdOpts.String())
}
