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

func TestTypes(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*handlers.ResourceHandler)(nil), &LxcHandler{})
	assert.Implements((*handlers.CLIHandler)(nil), &LxcHandler{})
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

const jsonLxcTemplate2 = `{
	"type": "vm/lxc",
	"lxc_template": {
		"download": {
			"distro": "ubuntu",
			"release": "xenial"
		}
	},
	"authentication_type":"none"
}`

const jsonLxcTemplate3 = `{
	"type": "vm/lxc",
	"lxc_template": {
		"download": {
			"distro": "ubuntu",
			"release": "xenial"
		}
	},
	"authentication_type":"pub_key",
	"ssh_public_key":"./ssh/rsa_pub"
}`

const jsonLxcTemplate4 = `{
	"type": "vm/lxc",
	"lxc_template": {
		"download": {
			"distro": "ubuntu",
			"release": "xenial"
		}
	},
	"authentication_type":1
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
	assert.Equal(model.LxcTemplate_NONE, modellxc.AuthenticationType, "none")

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate2).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
	assert.Equal(model.LxcTemplate_NONE, modellxc.AuthenticationType, "none")

	// m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate3).Bytes())
	// assert.NoError(err)
	// assert.IsType((*model.LxcTemplate)(nil), m)
	// modellxc = m.(*model.LxcTemplate)
	// assert.Nil(modellxc.GetLxcImage())
	// assert.NotNil(modellxc.GetLxcTemplate())
	// assert.Equal(model.LxcTemplate_PUB_KEY, modellxc.AuthenticationType, "pub_key")
	// assert.NotEmpty(modellxc.SshPublicKey)

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate4).Bytes())
	// assert.EqualError(err,"ssh_public_key is not set")
}
