package qemu

import (
	"bytes"
	"testing"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestResourceName(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("vm/qemu", handlers.ResourceName(&QemuHandler{}))
}

func TestTypes(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*handlers.ResourceHandler)(nil), &QemuHandler{})
	assert.Implements((*handlers.CLIHandler)(nil), &QemuHandler{})
}

const jsonQemuImage1 = `{
	"type": "vm/qemu",
	"qemu_image": {
		"download_url": "http://example.com/"
	}
}`


func TestQemuHandler_ParseTemplate(t *testing.T) {
	assert := assert.New(t)
	h := &QemuHandler{}
	m, err := h.ParseTemplate(bytes.NewBufferString(jsonQemuImage1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.QemuTemplate)(nil), m)
	modelqemu := m.(*model.QemuTemplate)
	assert.NotNil(modelqemu.GetQemuImage())
}
