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
	"ssh_public_key":"ssh-rsa AAAA"
}`

const jsonLxcTemplate4 = `{
	"type": "vm/lxc",
	"lxc_template": {
		"download": {
			"distro": "ubuntu",
			"release": "xenial"
		}
	},
	"authentication_type":"pub_key",
	"ssh_public_key":"./ssh/radf"
}`

const margeJson1 = `{"authentication_type":"none"}`
const margeJson2 = `{"authentication_type":"pub_key","ssh_public_key":"ssh-rsa AAAA"}`
const margeJson3 = `{"authentication_type":"pub_key","ssh_public_key":""}`

func TestLxcHandler_ParseTemplate(t *testing.T) {
	assert := assert.New(t)
	h := &LxcHandler{}
	m, err := h.ParseTemplate(bytes.NewBufferString(jsonLxcImage1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc := m.(*model.LxcTemplate)
	assert.NotNil(modellxc.GetLxcImage())
	assert.Equal(modellxc.GetLxcImage().DownloadUrl, "http://example.com/")
	assert.Nil(modellxc.GetLxcTemplate())

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
	assert.Equal(modellxc.GetLxcTemplate().Template, "download")
	assert.Equal(modellxc.GetLxcTemplate().Distro, "ubuntu")
	assert.Equal(modellxc.GetLxcTemplate().Release, "xenial")
	assert.Equal(model.AuthenticationType_NONE, modellxc.AuthenticationType, "none")

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate2).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
	assert.Equal(modellxc.GetLxcTemplate().Template, "download")
	assert.Equal(modellxc.GetLxcTemplate().Distro, "ubuntu")
	assert.Equal(modellxc.GetLxcTemplate().Release, "xenial")
	assert.Equal(model.AuthenticationType_NONE, modellxc.AuthenticationType, "none")

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate3).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
	assert.Equal(modellxc.GetLxcTemplate().Template, "download")
	assert.Equal(modellxc.GetLxcTemplate().Distro, "ubuntu")
	assert.Equal(modellxc.GetLxcTemplate().Release, "xenial")
	assert.Equal(model.AuthenticationType_PUB_KEY, modellxc.AuthenticationType, "pub_key")
	assert.NotEmpty(modellxc.SshPublicKey)

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate4).Bytes())
	assert.EqualError(err, "Invalid template vm/lxc: ssh_public_key is invalid")
}

func TestLxcHandler_MargeJSON(t *testing.T) {
	assert := assert.New(t)
	h := &LxcHandler{}
	var dest model.ResourceTemplate = &model.LxcTemplate{}

	err := h.MergeJSON(dest, bytes.NewBufferString(margeJson1).Bytes()) // instance_id := strings.TrimSpace(stdout.String())
	d := dest.(*model.LxcTemplate)
	assert.Nil(err)
	assert.IsType((*model.LxcTemplate)(nil), dest)
	assert.Equal(d.AuthenticationType, model.AuthenticationType_NONE)

	dest = &model.LxcTemplate{}
	err = h.MergeJSON(dest, bytes.NewBufferString(margeJson2).Bytes())
	d = dest.(*model.LxcTemplate)
	assert.Nil(err)
	assert.IsType((*model.LxcTemplate)(nil), dest)
	assert.Equal(d.AuthenticationType, model.AuthenticationType_PUB_KEY)
	assert.Equal(d.SshPublicKey, "ssh-rsa AAAA")

	dest = &model.LxcTemplate{}
	err = h.MergeJSON(dest, bytes.NewBufferString(margeJson3).Bytes())
	d = dest.(*model.LxcTemplate)
	assert.EqualError(err, "Invalid template vm/lxc: ssh_public_key is not set")
}
