// +build terraform

package openvdc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var getInterfaces = func() interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":     "veth",
			"ipv4addr": "0.0.0.0",
			"gateway":  "0.0.0.1",
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

func TestRenderInterfaceOpt(t *testing.T) {
	assert := assert.New(t)

	parsedInterfaces, err := renderInterfaceOpt(getInterfaces)
	assert.NoError(err)
	assert.Equal(string(parsedInterfaces), `"interfaces":[{"gateway":"0.0.0.1","ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]`)

	var buf bytes.Buffer
	// buf.WriteString("{")
	// buf.Write(parsedInterfaces)
	// buf.WriteString("}")

	buf.WriteString(`{"interfaces":[{"gateway":"0.0.0.1","ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]}`)

	var interface_template struct {
		Interfaces map[string]json.RawMessage `json:"interfaces,omitempty"`
	}

	// {
	// 	gateway string      `json:"gateway,omitempty"`
	// 	ipv4addr string     `json:"ipv4addr,omitempty"`
	// 	ifType string         `json:"type,omitempty"`
	// }
	fmt.Println(buf.String())
	json.Unmarshal(buf.Bytes(), &interface_template)
	fmt.Println(interface_template.Interfaces)
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
	assert.Equal(cmdOpts.String(), `{"memory_gb":1024,"vcpu":1,"interfaces":[{"gateway":"0.0.0.1","ipv4addr":"0.0.0.0","type":"veth"},{"ipv4addr":"1.0.0.0","type":"veth"}]}`)
}
