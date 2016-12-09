package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalRegistry(t *testing.T) {
	assert := assert.New(t)
	reg := NewLocalRegistry()
	assert.Implements((*TemplateFinder)(nil), reg)
}

func TestLocalRegistryFind(t *testing.T) {
	assert := assert.New(t)
	reg := NewLocalRegistry()
	rt, err := reg.Find("../templates/centos/7/lxc.json")
	assert.NoError(err)
	assert.NotNil(rt)
	assert.Equal(rt.source, reg)
}
