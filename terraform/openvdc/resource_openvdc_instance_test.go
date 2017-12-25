// +build terraform

package openvdc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var getInterfaces = func() interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":     "veth",
			"ipv4addr": "0.0.0.0",
			"ipv4gateway":  "0.0.0.1",
		},
		map[string]interface{}{
			"type":     "veth",
			"ipv4addr": "1.0.0.0",
		},
	}
}

var getResources = func() interface{} {
	return map[string]interface{}{
		"vcpu":      1,
		"memory_gb": 1024,
	}
}

func TestRenderListOpt(t *testing.T) {
	assert := assert.New(t)

	parsedInterfaces, err := renderListOpt("interfaces", getInterfaces())
	assert.NoError(err)
	assert.Equal(string(parsedInterfaces), `"interfaces":[{"ipv4gateway":"0.0.0.1","ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]`)
}

func TestRenderResourceOpt(t *testing.T) {
	assert := assert.New(t)

	parsedResources, err := renderMapOpt(getResources())
	assert.NoError(err)
	assert.Equal(string(parsedResources), `"memory_gb":1024,"vcpu":1`)
}

func TestRenderCmdOpt(t *testing.T) {
	assert := assert.New(t)

	p := []option{
		func() ([]byte, error) { return renderMapOpt(getResources()) },
		func() ([]byte, error) { return renderListOpt("interfaces", getInterfaces()) },
	}
	cmdOpts, err := renderCmdOpt(p)
	assert.NoError(err)
	assert.Equal(cmdOpts.String(), `{"memory_gb":1024,"vcpu":1,"interfaces":[{"ipv4gateway":"0.0.0.1","ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]}`)
}
