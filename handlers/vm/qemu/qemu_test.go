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
		"download_url": "http://example.com/",
		"format": "raw"
	}
}`

const jsonQemuImage2 = `{
	"type": "vm/qemu",
	"qemu_image": {
		"download_url": "http://example.com/",
		"format": "raw"
	},
	"authentication_type":"pub_key",
	"ssh_public_key":"ssh-rsa AAAA"
}`

const jsonQemuImage3 = `{
	"type": "vm/qemu",
	"qemu_image": {
		"download_url": "http://example.com/",
		"format": "raw"
	},
	"authentication_type":"none",
	"ssh_public_key":"ssh-rsa AAAA"
}`

const margeJson1 = `{"authentication_type":"none"}`
const margeJson2 = `{"authentication_type":"pub_key","ssh_public_key":"ssh-rsa AAAA"}`
const margeJson3 = `{"authentication_type":"pub_key","ssh_public_key":""}`

func TestQemuHandler_ParseTemplate(t *testing.T) {
	assert := assert.New(t)
	h := &QemuHandler{}
	m, err := h.ParseTemplate(bytes.NewBufferString(jsonQemuImage1).Bytes())
	assert.NoError(err)
	assert.IsType((*model.QemuTemplate)(nil), m)
	modelqemu := m.(*model.QemuTemplate)
	assert.NotNil(modelqemu.GetQemuImage())
	assert.Equal(modelqemu.GetQemuImage().GetDownloadUrl(), "http://example.com/")
	assert.Equal(modelqemu.GetQemuImage().GetFormat().String(), "RAW")
	assert.Equal(model.AuthenticationType_NONE, modelqemu.AuthenticationType, "none")

	// Testing authentication_type and ssh_pub_key
	m, err = h.ParseTemplate(bytes.NewBufferString(jsonQemuImage2).Bytes())
	assert.NoError(err)
	assert.IsType((*model.QemuTemplate)(nil), m)
	modelqemu = m.(*model.QemuTemplate)
	assert.NotNil(modelqemu.GetQemuImage())
	assert.Equal(modelqemu.GetQemuImage().GetDownloadUrl(), "http://example.com/")
	assert.Equal(modelqemu.GetQemuImage().GetFormat().String(), "RAW")
	assert.Equal(model.AuthenticationType_PUB_KEY, modelqemu.AuthenticationType, "pub_key")
	assert.Equal(modelqemu.SshPublicKey, "ssh-rsa AAAA")

	m, err = h.ParseTemplate(bytes.NewBufferString(jsonQemuImage3).Bytes())
	assert.NoError(err)
	assert.IsType((*model.QemuTemplate)(nil), m)
	modelqemu = m.(*model.QemuTemplate)
	assert.NotNil(modelqemu.GetQemuImage())
	assert.Equal(modelqemu.GetQemuImage().GetDownloadUrl(), "http://example.com/")
	assert.Equal(modelqemu.GetQemuImage().GetFormat().String(), "RAW")
	assert.Equal(model.AuthenticationType_NONE, modelqemu.AuthenticationType, "none")
}

func TestQemuHandler_MergeArgs(t *testing.T) {
	assert := assert.New(t)
	h := &QemuHandler{}
	var dest model.ResourceTemplate = &model.QemuTemplate{}
	args := []string{`--authentication_type="none"`}
	err := h.MergeArgs(dest, args)
	d := dest.(*model.QemuTemplate)
	assert.Nil(err)
	assert.IsType((*model.QemuTemplate)(nil), dest)
	assert.Equal(model.AuthenticationType_NONE, d.AuthenticationType)

	dest = &model.QemuTemplate{}
	args = []string{"--vcpu=2"}
	err = h.MergeArgs(dest, args)
	d = dest.(*model.QemuTemplate)
	assert.Nil(err)
	assert.IsType((*model.QemuTemplate)(nil), dest)
	assert.Equal(2, int(d.GetVcpu()))

	dest = &model.QemuTemplate{}
	args = []string{`--authentication_type=pub_key`, `--ssh_public_key="ssh-rsa AAAA"`}
	err = h.MergeArgs(dest, args)
	d = dest.(*model.QemuTemplate)
	assert.Nil(err)
	assert.IsType((*model.QemuTemplate)(nil), dest)
	assert.Equal(model.AuthenticationType_PUB_KEY, d.AuthenticationType)
	assert.Equal("ssh-rsa AAAA", d.SshPublicKey)
}

func TestQemuHandler_MargeJSON(t *testing.T) {
	assert := assert.New(t)
	h := &QemuHandler{}
	var dest model.ResourceTemplate = &model.QemuTemplate{}

	err := h.MergeJSON(dest, bytes.NewBufferString(margeJson1).Bytes())
	d := dest.(*model.QemuTemplate)
	assert.Nil(err)
	assert.IsType((*model.QemuTemplate)(nil), dest)
	assert.Equal(d.AuthenticationType, model.AuthenticationType_NONE)

	dest = &model.QemuTemplate{}
	err = h.MergeJSON(dest, bytes.NewBufferString(margeJson2).Bytes())
	d = dest.(*model.QemuTemplate)
	assert.Nil(err)
	assert.IsType((*model.QemuTemplate)(nil), dest)
	assert.Equal(d.AuthenticationType, model.AuthenticationType_PUB_KEY)
	assert.Equal(d.SshPublicKey, "ssh-rsa AAAA")

	dest = &model.QemuTemplate{}
	err = h.MergeJSON(dest, bytes.NewBufferString(margeJson3).Bytes())
	d = dest.(*model.QemuTemplate)
	assert.EqualError(err, "Invalid template vm/qemu: ssh_public_key is not set")
}
