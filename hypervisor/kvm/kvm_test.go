// build +linux

package kvm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestProviderRegistration(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("kvm")
	assert.NotNil(p, "Check kvm provider is registered.")
	assert.Equal("kvm", p.Name())
	assert.Implements((*hypervisor.HypervisorProvider)(nil), p)
}

func TestKVMHypervisorProvider_CreateDriver(t *testing.T) {
	assert := assert.New(t)
	p, _ := hypervisor.FindProvider("kvm")

	d, err := p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, &model.KvmTemplate{})
	assert.NoError(err)
	assert.Implements((*hypervisor.HypervisorDriver)(nil), d)
	_, err = p.CreateDriver(&model.Instance{Id: "i-xxxxx"}, nil)
	assert.Error(err, "KVMHypvisorProvider.CreateDriver should fail if not with *model.KVMTemplate")
}
