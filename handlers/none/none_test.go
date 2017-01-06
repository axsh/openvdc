package none

import (
	"testing"

	"github.com/axsh/openvdc/handlers"
	"github.com/stretchr/testify/assert"
)

func TestResourceName(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("none", handlers.ResourceName(&NoneHandler{}))
}
