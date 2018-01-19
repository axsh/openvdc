package esxi

import (
	"fmt"
	"os"
	"testing"

	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestProviderRegistration(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("esxi")
	assert.NotNil(p, "Check if esxi provider is registered.")
	assert.Equal("esxi", p.Name())
	assert.Implements((*hypervisor.HypervisorProvider)(nil), p)
}

func TestEsxiHypervisorProvider_CreateDriver(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("esxi")

	d, err := p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, &model.EsxiTemplate{})
	assert.NoError(err)
	assert.Implements((*hypervisor.HypervisorDriver)(nil), d)
	_, err = p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, nil)
	assert.Error(err, "ESXIHypvisorProvider.CreateDriver should fail if not with *model.EsxiTemplate")
}

func TestCaptureStdout(t *testing.T) {
	assert := assert.New(t)
	stdout, err := captureStdout(func() error {
		fmt.Fprint(os.Stdout, "sample output")
		return nil
	})
	fmt.Println(stdout)
	assert.NoError(err)
	assert.Equal("sample output", string(stdout))
}
