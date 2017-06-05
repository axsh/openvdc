package lxc

import (
	"bytes"
	"testing"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestResourceName(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("vm/kvm", handlers.ResourceName(&KvmHandler{}))
}

func TestTypes(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*handlers.ResourceHandler)(nil), &KvmHandler{})
	assert.Implements((*handlers.CLIHandler)(nil), &KvmHandler{})
}

const jsonKvmImage1 = `{
	"type": "vm/kvm",
	"kvm_image": {
		"download_url": "http://example.com/"
	}
}`


func TestKvmHandler_ParseTemplate(t *testing.T) {
	assert := assert.New(t)
	h := &KvmHandler{}
	m, err := h.ParseTemplate(bytes.NewBufferString(jsonKvmImage1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.KvmTemplate)(nil), m)
	modelkvm := m.(*model.KvmTemplate)
	assert.NotNil(modelkvm.GetKvmImage())
	assert.Nil(modelkvm.GetKvmTemplate())
}
