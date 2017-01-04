package registry

import (
	"testing"

	"strings"

	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

const json1 = `
{
  "$schema": "https://raw.githubusercontent.com/axsh/openvdc/master/schema/v1.json#",
  "title": "CentOS7",
  "template": {
    "type": "vm/lxc",
    "lxc_image": {
      "download_url": "https://images.linuxcontainers.org/1.0/images/d767cfe9a0df0b2213e28b39b61e8f79cb9b1e745eeed98c22bc5236f277309a/export"
    }
  }
}`

func TestParseJSON(t *testing.T) {
	assert := assert.New(t)
	root, err := parseResourceTemplate(strings.NewReader(json1))
	assert.NoError(err)
	assert.NotNil(root)
	assert.Equal("CentOS7", root.Title)
	assert.IsType(new(model.LxcTemplate), root.Template)
	lxc := root.Template.(*model.LxcTemplate)
	assert.IsType(new(model.LxcTemplate_Image), lxc.GetLxcImage())
	assert.Equal("https://images.linuxcontainers.org/1.0/images/d767cfe9a0df0b2213e28b39b61e8f79cb9b1e745eeed98c22bc5236f277309a/export",
		lxc.GetLxcImage().GetDownloadUrl())
}
