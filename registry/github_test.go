package registry

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/stretchr/testify/assert"

	// To pass TestFind
	_ "github.com/axsh/openvdc/handlers/vm/lxc"
)

func init() {
	GithubDefaultRef = unittest.GithubDefaultRef
}

func TestNewGithubRegistry(t *testing.T) {
	assert := assert.New(t)
	reg := NewGithubRegistry("xxx")
	assert.Implements((*CachedRegistry)(nil), reg)
}

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
	assert.Equal(rt.source, reg)

	// Try finding unknown template name.
	rt, err = reg.Find("should-not-exist")
	assert.Error(err)
	assert.Nil(rt)
}
