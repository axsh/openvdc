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
