// +build terraform

package openvdc

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var getInterfaces = func() interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":        "veth",
			"ipv4addr":    "0.0.0.0",
			"ipv4gateway": "0.0.0.1",
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

type nic struct {
	Ipv4gateway string `json:"ipv4gateway,omitempty"`
	Ipv4addr    string `json:"ipv4addr,omitempty"`
	Type        string `json:"type,omitempty"`
}

func assertNics(assert *assert.Assertions, nics []nic) {

	assert.Len(nics, 2)

	assert.Equal("0.0.0.1", nics[0].Ipv4gateway)
	assert.Equal("0.0.0.0", nics[0].Ipv4addr)
	assert.Equal("veth", nics[0].Type)

	assert.Equal("1.0.0.0", nics[1].Ipv4addr)
	assert.Equal("veth", nics[1].Type)
	assert.Empty(nics[1].Ipv4gateway)
}

func TestRenderListOpt(t *testing.T) {
	assert := assert.New(t)
	parsedInterfaces, err := renderListOpt("interfaces", getInterfaces())
	assert.NoError(err)

	// dirty workaround to force a json output for easier parsing
	parsedInterfaces = append([]byte("{"), parsedInterfaces...)
	parsedInterfaces = append(parsedInterfaces, []byte("}")[0])

	var data struct {
		Nics []nic `json:"interfaces,omitempty"`
	}
	if err := json.Unmarshal(parsedInterfaces, &data); err != nil {
		log.Fatalf("failed json.Unmarshal: %v", err)
	}
	assertNics(assert, data.Nics)
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

	var opts struct {
		Memory int   `json:"memory_gb,omitempty"`
		Vcpu   int   `json:"vcpu,omitempty"`
		Nics   []nic `json:"interfaces,omitempty"`
	}

	if err := json.Unmarshal(cmdOpts.Bytes(), &opts); err != nil {
		log.Fatalf("failed json.Unmarshal: %v", err)
	}
	assert.Equal(1024, opts.Memory)
	assert.Equal(1, opts.Vcpu)
	assertNics(assert, opts.Nics)
}
