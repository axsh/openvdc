package null

import (
	"testing"

	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestProviderRegistration(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("null")
	assert.NotNil(p, "Check null provider is registered.")
	assert.Equal("null", p.Name())
	assert.Implements((*hypervisor.HypervisorProvider)(nil), p)
}

func TestNullHypervisorProvider_CreateDriver(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("null")

	d, err := p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, &model.NullTemplate{})
	assert.NoError(err)
	assert.Implements((*hypervisor.HypervisorDriver)(nil), d)
	_, err = p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, nil)
	assert.Error(err, "NullHypvisorProvider.CreateDriver should fail if not with *model.NullTemplate")
}
