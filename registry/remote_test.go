package registry

import (
	"fmt"
	"testing"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/stretchr/testify/assert"
)

func TestNewRemoteRegistry(t *testing.T) {
	assert := assert.New(t)
	reg := NewRemoteRegistry()
	assert.Implements((*TemplateFinder)(nil), reg)
}

func TestRemoteRegistryFind(t *testing.T) {
	assert := assert.New(t)
	reg := NewRemoteRegistry()
	rt, err := reg.Find(fmt.Sprintf("%s/%s/%s/templates/centos/7/lxc.json",
		githubRawURI, githubRepoSlug, unittest.GithubDefaultRef))
	assert.NoError(err)
	assert.NotNil(rt)
	assert.Equal(rt.source, reg)
}
