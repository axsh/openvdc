package registry

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	_, err = os.Stat(filepath.Join(dir, "registry", "github.com-axsh-openvdc-images", reg.Branch))
	assert.NoError(err)
	_, err = os.Stat(filepath.Join(dir, "registry", "github.com-axsh-openvdc-images", reg.Branch+".sha"))
	assert.NoError(err)
	c, err := exec.Command("find", dir).Output()
	fmt.Println(dir)
	fmt.Println(string(c))
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
	// Try finding existing image name.
	mi, err := reg.Find("centos-7")
	assert.NoError(err)
	assert.NotNil(mi)

	// Try finding unknown image name.
	mi, err = reg.Find("should-not-exist")
	assert.Error(err)
	assert.Nil(mi)
}
