package registry

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	gitDefaultURI = "https://github.com/axsh/openvdc.git"
	gitDefaultRef = "master"
)

type GitRegistry struct {
	GitCmd                  string
	confDir                 string
	Ref                     string
	URI                     string
	SubtreePath             string
	ForceCheckToRemoteAfter time.Duration
	FetchRetry              int
	Depth                   int
}

func NewGitRegistry(confDir string, repoURI string) *GitRegistry {
	return &GitRegistry{
		GitCmd:                  "git",
		confDir:                 confDir,
		Ref:                     gitDefaultRef,
		URI:                     repoURI,
		SubtreePath:             ".",
		ForceCheckToRemoteAfter: 12 * time.Hour,
		FetchRetry:              3,
		Depth:                   10,
	}
}

func (r *GitRegistry) String() string {
	return fmt.Sprintf("%s", r.URI)
}

func (r *GitRegistry) LocateURI(name string) string {
	if !strings.HasSuffix(name, ".json") {
		name += ".json"
	}
	return fmt.Sprintf("%s/%s", r.URI, name)
}

// Find queries resource template details from local registry cache.
func (r *GitRegistry) Find(templateName string) (*RegistryTemplate, error) {
	f, err := r.openCached(templateName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tmpl, err := parseResourceTemplate(f)
	if err != nil {
		return nil, err
	}
	rt := &RegistryTemplate{
		Name:     templateName,
		source:   r,
		Template: tmpl,
	}
	return rt, nil
}

func (r *GitRegistry) LoadRaw(templateName string) ([]byte, error) {
	buf := new(bytes.Buffer)
	f, err := r.openCached(templateName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := io.Copy(buf, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ValidateCache validates the local cache folder items.
func (r *GitRegistry) ValidateCache() bool {
	stat, err := os.Stat(r.localCachePath())
	return err == nil && stat.IsDir()
}

func (r *GitRegistry) localCachePath() string {
	// $confDir/registry/git-${fqdn_host}-${path}/${ref}
	url, err := url.Parse(r.URI)
	if err != nil {
		panic(err)
	}
	repo_path := url.EscapedPath()
	if ext := filepath.Ext(repo_path); ext != "" {
		repo_path = repo_path[:len(repo_path)-len(ext)]
	}
	return filepath.Join(r.confDir, "registry",
		fmt.Sprintf("git-%s%s", url.Hostname(), strings.Replace(repo_path, "/", "-", -1)))
}

func (r *GitRegistry) openCached(templateName string) (*os.File, error) {
	if !r.ValidateCache() {
		return nil, ErrLocalCacheNotReady
	}
	json_path := filepath.Clean(filepath.Join(r.localCachePath(), filepath.FromSlash(r.SubtreePath), filepath.FromSlash(templateName)+".json"))
	f, err := os.Open(json_path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrUnknownTemplateName
		}
		return nil, err
	}
	return f, nil
}

func (r *GitRegistry) IsCacheObsolete() (bool, error) {
	if !r.ValidateCache() {
		// There was no local cache.
		return true, nil
	}
	return false, nil
}

func (r *GitRegistry) Fetch() error {
	_, err := os.Stat(r.localCachePath())
	if os.IsNotExist(err) {
		// git clone
		basedir, gitdir := filepath.Split(r.localCachePath())
		_, err := os.Stat(basedir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(basedir, 755); err != nil {
				return errors.Wrap(err, "Failed MkdirAll")
			}
		}

		git_clone := exec.Command(r.GitCmd, "clone", "--depth", strconv.Itoa(r.Depth), "--branch", r.Ref, "--config", "core.sparsecheckout=true", "--no-checkout", r.URI, gitdir)
		git_clone.Dir = basedir
		if err := git_clone.Run(); err != nil {
			errors.Wrap(err, "Failed git clone")
		}
		println(r.GitCmd, "clone", "--depth", strconv.Itoa(r.Depth), "--branch", r.Ref, "--config", "core.sparsecheckout=true", "--no-checkout", r.URI, gitdir)
		if r.SubtreePath != "." {
			f, err := os.OpenFile(filepath.Join(r.localCachePath(), ".git", "info", "sparse-checkout"), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return errors.Wrap(err, "Failed to create .git/info/sparse-checkout")
			}
			fmt.Fprintln(f, r.SubtreePath)
			f.Close()
		}
		git_chkout := exec.Command(r.GitCmd, "checkout")
		git_chkout.Dir = r.localCachePath()
		if err := git_chkout.Run(); err != nil {
			return errors.Wrap(err, "Failed git checkout")
		}
	} else {
		// git pull
		git_pull := exec.Command(r.GitCmd, "pull")
		git_pull.Dir = r.localCachePath()
		if err := git_pull.Run(); err != nil {
			return errors.Wrap(err, "Failed git pull")
		}
	}
	return nil
}
