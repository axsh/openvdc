package registry

import (
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

const validJSON1 = `
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
const invalidJSON1 = `
{
  "$schema": "https://raw.githubusercontent.com/axsh/openvdc/master/schema/v1.json#",
  "title": "CentOS7",
  "template": {
  }
}`

func TestValidateTemplate(t *testing.T) {
	assert := assert.New(t)
	var err error

	err = ValidateTemplate([]byte(validJSON1))
	assert.NoError(err)
	err = ValidateTemplate([]byte(invalidJSON1))
	if err != nil {
		if err, ok := err.(*ErrInvalidTemplate); ok {
			err.Dump(os.Stderr)
		}
	}
	assert.Error(err)
}
