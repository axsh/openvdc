package registry

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	githubURI             = "https://github.com"
	githubRawURI          = "https://raw.githubusercontent.com"
	githubRepoSlug        = "axsh/openvdc"
	defaultPath           = "templates"
	mimeTypeGitUploadPack = "application/x-git-upload-pack-advertisement"
)

// Buildtime constant
var GithubDefaultRef = "master"

type GithubRegistry struct {
	confDir                 string
	Branch                  string
	RepoSlug                string
	TreePath                string
	ForceCheckToRemoteAfter time.Duration
	FetchRetry              int
}

func NewGithubRegistry(confDir string) *GithubRegistry {
	return &GithubRegistry{
		confDir:                 confDir,
		Branch:                  GithubDefaultRef,
		RepoSlug:                githubRepoSlug,
		TreePath:                defaultPath,
		ForceCheckToRemoteAfter: 1 * time.Hour,
		FetchRetry:              3,
	}
}

func (r *GithubRegistry) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", githubRawURI, r.RepoSlug, r.Branch, r.TreePath)
}

func (r *GithubRegistry) LocateURI(name string) string {
	if !strings.HasSuffix(name, ".json") {
		name += ".json"
	}
	return fmt.Sprintf("%s/%s/%s/%s", githubRawURI, r.RepoSlug, r.Branch, name)
}

func (r *GithubRegistry) findRemoteRef() (ref *gitRef, err error) {
	refs, err := gitLsRemote(r.RepoSlug)
	if err != nil {
		return
	}
	ref = findRef(refs, r.Branch)
	return
}

func (r *GithubRegistry) localCachePath() string {
	// $confDir/registry/${fqdn_host}-${user}-${repo}/${ref}
	return filepath.Join(r.confDir, "registry",
		fmt.Sprintf("github.com-%s", strings.Replace(r.RepoSlug, "/", "-", 1)),
		r.Branch)
}

// Find queries resource template details from local registry cache.
func (r *GithubRegistry) Find(templateName string) (*RegistryTemplate, error) {
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

func (r *GithubRegistry) LoadRaw(templateName string) ([]byte, error) {
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

func (r *GithubRegistry) openCached(templateName string) (*os.File, error) {
	if !r.ValidateCache() {
		return nil, ErrLocalCacheNotReady
	}
	f, err := os.Open(filepath.Join(r.localCachePath(), templateName+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrUnknownTemplateName
		}
		return nil, err
	}
	return f, nil
}

// ValidateCache validates the local cache folder items.
func (r *GithubRegistry) ValidateCache() bool {
	stat, err := os.Stat(r.localCachePath())
	// TODO: more check for .sha file.
	return err == nil && stat.IsDir()
}

// IsCacheObsolete consults to the remote machine image registry and checks if
// the local cache is obsolete.
func (r *GithubRegistry) IsCacheObsolete() (bool, error) {
	if !r.ValidateCache() {
		// There was no local cache.
		return true, nil
	}
	shaPath := r.localCachePath() + ".sha"
	stat, err := os.Stat(shaPath)
	if err != nil {
		return false, err
	}
	if time.Since(stat.ModTime()) < r.ForceCheckToRemoteAfter {
		// Skip to make a remote request to check version.
		return false, nil
	}
	shabuf, err := ioutil.ReadFile(shaPath)
	if err != nil {
		return false, err
	}
	sha := strings.TrimSpace(string(shabuf))

	ref, err := r.findRemoteRef()
	if err != nil {
		return false, err
	}
	// Refresh timestamp
	os.Truncate(shaPath, stat.Size())
	return (ref.Sha != sha), nil
}

// Fetch download and extract the remote image registry to local folder.
func (r *GithubRegistry) Fetch() error {
	ref, err := r.findRemoteRef()
	if err != nil {
		return err
	}
	if ref == nil {
		return fmt.Errorf("Counld not find the branch: %s", r.Branch)
	}

	tmpDest, err := ioutil.TempDir("", "gh-images-reg")
	defer func() {
		err := os.RemoveAll(tmpDest)
		if err != nil {
			log.WithError(err).Errorf("Failed to cleanup tmp directory: %s", tmpDest)
		}
	}()

	// Create local registry cache.
	regDir := r.localCachePath()
	if _, err = os.Stat(regDir); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(regDir), 0755)
		if err != nil {
			return err
		}
	} else {
		// Clean exisiting cache
		err = os.RemoveAll(regDir)
		if err != nil {
			return err
		}
	}

	// https://github.com/axsh/openvdc/archive/%{sha}.zip
	zipLinkURI := fmt.Sprintf("%s/%s/archive/%s.zip", githubURI, r.RepoSlug, ref.Sha)
	tmpzip, err := ioutil.TempFile(tmpDest, "zip")
	if err != nil {
		return err
	}
	defer func() {
		tmpzip.Close()
		os.Remove(tmpzip.Name())
	}()

	var res *http.Response
	for i := 0; i < r.FetchRetry; i++ {
		res, err = http.Get(zipLinkURI)
		if err == nil {
			break
		}
		log.WithError(err).Warnf("http.Get failed retrying... %d/%d", i+1, r.FetchRetry)
	}
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to request %s with %s", zipLinkURI, res.Status)
	}

	_, err = io.Copy(tmpzip, res.Body)
	if err != nil {
		// TODO: retry fetching because it might be a network error.
		return err
	}

	prefix := filepath.Join(
		fmt.Sprintf("%s-%s", path.Base(r.RepoSlug), ref.Sha),
		filepath.FromSlash(r.TreePath))
	// The zip file contains a root folder with "%{repo}-%{sha}" convention.
	// "openvdc-e77ed15f3b2ba582087afa226ace61a6756f65dd/templates"
	// Extract the archive to regDir.
	err = unzip(tmpzip.Name(), regDir, prefix)
	if err != nil {
		return err
	}
	// Save current sha
	return ioutil.WriteFile(regDir+".sha", []byte(ref.Sha), 0644)
}

// http://blog.ralch.com/tutorial/golang-working-with-zip/
func unzip(archive, target string, prefix string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	atime := time.Now()
	for _, file := range reader.File {
		if prefix != "" && !strings.HasPrefix(filepath.FromSlash(file.Name), prefix) {
			continue
		}
		// Slide path in .zip to the length of prefix.
		path := filepath.Join(target, filepath.FromSlash(file.Name[len(prefix)+1:]))
		if path == "" {
			continue
		}
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, file.Mode())
			if err != nil {
				return err
			}
			os.Chtimes(path, atime, file.ModTime())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
		os.Chtimes(path, atime, file.ModTime())
	}

	return nil
}

type gitRef struct {
	Sha  string
	Ref  string
	Caps []string
}

func gitLsRemote(repoSlug string) (refs []*gitRef, err error) {
	// https://github.com/git/git/blob/master/Documentation/technical/http-protocol.txt
	// Build git-upload-pack smart request
	uri := fmt.Sprintf("%s/%s.git/info/refs?service=git-upload-pack", githubURI, repoSlug)
	r, err := http.Get(uri)
	if err != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %s", uri, r.Status)
	}
	if r.Header.Get("Content-Type") != mimeTypeGitUploadPack {
		return nil, fmt.Errorf("Invalid content type: %s", r.Header.Get("Content-Type"))
	}

	pktLine := func(in io.Reader) (string, error) {
		var l int
		var err error

		b := make([]byte, 4)
		l, err = in.Read(b)
		if err != nil {
			return "", err
		}
		if l != len(b) {
			return "", fmt.Errorf("Invalid pkt-line header")
		}
		blen, err := strconv.ParseUint(string(b), 16, 32)
		if err != nil {
			return "", err
		}
		if blen == 0 {
			// "0000" terminal magic. treat as non-error.
			return "0000", nil
		}
		b2 := make([]byte, blen-4)
		l, err = in.Read(b2)
		if err != nil {
			return "", err
		}
		if l != len(b2) {
			return "", fmt.Errorf("Invalid pkt-line content length: read=%d, expected=%d", l, len(b2))
		}
		// dispose last "\n"
		return string(b2[:l-1]), nil
	}

	parseLine := func(s string) (sha string, ref string, caps []string) {
		sp1 := strings.SplitN(s, " ", 2)
		sha = sp1[0]
		l := strings.IndexByte(sp1[1], 0) // Find \0 (NUL)
		if l == -1 {
			ref = sp1[1]
			return
		}
		ref = sp1[1][0:l]
		caps = strings.Split(sp1[1][l+1:], " ")
		return
	}

	defer r.Body.Close()
	// smart_reply
	smartReply, err := pktLine(r.Body)
	if smartReply != "# service=git-upload-pack" {
		return nil, fmt.Errorf("Invalid smart_reply: '%s'", smartReply)
	}
	blank, err := pktLine(r.Body)
	if blank != "0000" {
		return nil, fmt.Errorf("Invalid sequence")
	}

	// ref_list
	for s, err := pktLine(r.Body); s != "0000" && err == nil; s, err = pktLine(r.Body) {
		sha, ref, caps := parseLine(s)
		refs = append(refs, &gitRef{Sha: sha, Ref: ref, Caps: caps})
	}
	return
}

func findRef(refs []*gitRef, name string) *gitRef {
	for _, r := range refs {
		if strings.HasSuffix(r.Ref, "/"+name) {
			return r
		}
	}
	return nil
}
