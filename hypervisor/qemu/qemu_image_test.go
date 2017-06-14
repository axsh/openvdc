
package qemu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImage(t *testing.T) {
	assert := assert.New(t)
	img := NewImage("path", "format")
	assert.NotNil(img)
	assert.Equal("format", img.Format)
	assert.Equal("path", img.Path)
}
