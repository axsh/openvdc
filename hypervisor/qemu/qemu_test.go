// build +linux

package qemu

import (
	"testing"

	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestProviderRegistration(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("qemu")
	assert.NotNil(p, "Check qemu provider is registered.")
	assert.Equal("qemu", p.Name())
	assert.Implements((*hypervisor.HypervisorProvider)(nil), p)
}

func TestQEMUHypervisorProvider_CreateDriver(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("qemu")

	d, err := p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, &model.QemuTemplate{})
	assert.NoError(err)
	assert.Implements((*hypervisor.HypervisorDriver)(nil), d)
	_, err = p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, nil)
	assert.Error(err, "QEMUHypvisorProvider.CreateDriver should fail if not with *model.QEMUTemplate")
}
