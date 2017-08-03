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

const jsonLxcTemplate5 = `{
	"type": "vm/lxc",
	"lxc_template": {
		"download": {
			"distro": "ubuntu",
			"release": "xenial"
		}
	},
	"authentication_type":"pub_key",
	"ssh_public_key":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDJNCHA2x1bX1S3vmudlHEYDalDfUfeXn/Ca/l4V9phaGA6R4R+MPYkkA4opctEOeUot5rZDhnXGIOmIxvZdfsaZf7ib0zUTSF9iO3H3KeP9jm0DauxbfFVufa6EqJHKzX9sR9hnMQTFenac/bXlfwJVjt6Xz4gGCvfQ2tY4zJPCy4C/tGjZfsD+dya2UVqWp7Sg0I3+iuvREGoeT/9ayKoOj/j8DXbKchjspm3JHcsgp5lTDxrnfiWgV1HrKeWhiKPaNKY70TJHDhsaUL7CLbtT/RHogRkbPAaiBdm5wxdvC37ziflsgsLX9cRpFNuBD3xeckRX2+QsKzGLa8Wp1T1XRdcekoVLCT6RBcx1hbawrxBb3M2PrKXMkbTg96TlrAIMtbpM1oMV5NhWJFe3Y6nET+6Z5j4TBuv3HN69FlGlWcl/+TNppuk3iJC/fAMOmxNyuhA1U6k/s0od3MbagPXmso9YkH9fhtuDaerv23hf7m68oDaz2nK/zfK47Bn+06tjpznR0XFwYK4Bhp2UCoXOFshBkHbpqnZupPcLd/dHSRDfXgOKTfptGvGAz7vwINXBAhPEc0G9GGnha0RTRct3hkrkUqkLS/0d05UXxeS6VyB0CJpDtdU8CXc5wyas+oelUYLOOdeCPnsYMIOGILrxBFD23GIQ6l9UWPseDc3Yw=="
}`

const margeJson1 = `{"authentication_type":"none"}`

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
	assert.Equal(model.LxcTemplate_NONE, modellxc.AuthenticationType, "none")

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate2).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
	assert.Equal(modellxc.GetLxcTemplate().Template, "download")
	assert.Equal(modellxc.GetLxcTemplate().Distro, "ubuntu")
	assert.Equal(modellxc.GetLxcTemplate().Release, "xenial")
	assert.Equal(model.LxcTemplate_NONE, modellxc.AuthenticationType, "none")

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate3).Bytes())
	assert.NoError(err)
	assert.IsType((*model.LxcTemplate)(nil), m)
	modellxc = m.(*model.LxcTemplate)
	assert.Nil(modellxc.GetLxcImage())
	assert.NotNil(modellxc.GetLxcTemplate())
	assert.Equal(modellxc.GetLxcTemplate().Template, "download")
	assert.Equal(modellxc.GetLxcTemplate().Distro, "ubuntu")
	assert.Equal(modellxc.GetLxcTemplate().Release, "xenial")
	assert.Equal(model.LxcTemplate_PUB_KEY, modellxc.AuthenticationType, "pub_key")
	assert.NotEmpty(modellxc.SshPublicKey)

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonLxcTemplate4).Bytes())
	assert.Error(err)
}

func TestLxcHandler_MargeJSON(t *testing.T) {
	assert := assert.New(t)
	h := &LxcHandler{}
	lxcTmpl := &model.LxcTemplate{}
	var dest model.ResourceTemplate = lxcTmpl

	//dest = model.LxcTemplate{}
	err := h.MergeJSON(dest, bytes.NewBufferString(jsonLxcImage1).Bytes()) // instance_id := strings.TrimSpace(stdout.String())
	assert.Nil(err)
	assert.IsType((*model.LxcTemplate)(nil), dest)
}
