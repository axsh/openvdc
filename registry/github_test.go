package registry

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGithubFetch(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "reg-test")
	assert.NoError(err, "Could not create temp dir for testing.")
	defer func() {
		os.RemoveAll(dir)
	}()
	reg := NewGithubRegistry(dir)
	err = reg.Fetch()
	assert.NoError(err)
	_, err = os.Stat(filepath.Join(dir, "registry", "github.com-axsh-openvdc", reg.Branch))
	assert.NoError(err)
	_, err = os.Stat(filepath.Join(dir, "registry", "github.com-axsh-openvdc", reg.Branch+".sha"))
	assert.NoError(err)
}

func TestGitLsRemote(t *testing.T) {
	assert := assert.New(t)
	refs, err := gitLsRemote(githubRepoSlug)
	assert.NoError(err)
	assert.NotNil(findRef(refs, "master"))
}

func TestFind(t *testing.T) {
	assert := assert.New(t)
	dir, err := ioutil.TempDir("", "reg-test")
	assert.NoError(err, "Could not create temp dir for testing.")
	defer func() {
		os.RemoveAll(dir)
	}()

	reg := NewGithubRegistry(dir)
	err = reg.Fetch()
	assert.NoError(err)
	// Try finding existing template name.
	rt, err := reg.Find("centos/7/lxc")
	assert.NoError(err)
	assert.NotNil(rt)

	// Try finding unknown template name.
	rt, err = reg.Find("should-not-exist")
	assert.Error(err)
	assert.Nil(rt)
}
