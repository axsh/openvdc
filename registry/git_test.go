package registry

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitFetch(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "reg-test")
	assert.NoError(err, "Could not create temp dir for testing.")
	defer func() {
		os.RemoveAll(dir)
	}()
	reg := NewGitRegistry(dir, "https://github.com/axsh/openvdc.git")
	// git clone happens at the first time.
	err = reg.Fetch()
	assert.NoError(err)
	_, err = os.Stat(filepath.Join(dir, "registry", "git-github.com-axsh-openvdc.git"))

	// Try second time then for git pull with no error.
	err = reg.Fetch()
	assert.NoError(err)
}

func TestGitFind(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "reg-test")
	assert.NoError(err, "Could not create temp dir for testing.")
	defer func() {
		os.RemoveAll(dir)
	}()

	reg := NewGitRegistry(dir, "https://github.com/axsh/openvdc.git")
	reg.SubtreePath = "templates"
	err = reg.Fetch()
	assert.NoError(err)
	// Try finding existing template name.
	rt, err := reg.Find("centos/7/lxc")
	assert.NoError(err)
	assert.NotNil(rt)
	assert.Equal(rt.source, reg)

	// Try finding unknown template name.
	rt, err = reg.Find("should-not-exist")
	assert.Error(err)
	assert.Nil(rt)

	// Confirm to fail with unknown subtree path.
	reg.SubtreePath = "nosuchdir"
	rt, err = reg.Find("centos/7/lxc")
	assert.Error(err)
	assert.Nil(rt)
}
