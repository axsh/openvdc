package lxc

import (
	"testing"

	"github.com/axsh/openvdc/handlers"
	"github.com/stretchr/testify/assert"
)

func TestResourceName(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("vm/lxc", handlers.ResourceName(&LxcHandler{}))
}
