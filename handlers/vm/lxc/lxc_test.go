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
	assert.Equal("vm/lxc", handlers.ResourceName(&LxcHandler{}))
}

const jsonLxcImage1 = `{
	"type": "vm/lxc",
	"lxc_image": {
		"download_url": "http://example.com/"
	}
}`

const jsonLxcTemplate1 = `{
	"type": "vm/lxc",
	"lxc_template": {
		"download": {
			"distro": "ubuntu",
			"release": "xenial"
		}
	}
}`

func TestLxcHandler_ParseTemplate(t *testing.T) {
	assert := assert.New(t)
	h := &LxcHandler{}
	m, err := h.ParseTemplate(bytes.NewBufferString(jsonLxcImage1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc := m.(*model.LxcTemplate)
	assert.NotNil(modellxc.GetLxcImage())
	assert.Nil(modellxc.GetLxcTemplate())

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
}
