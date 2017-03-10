package registry

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

// Handle resource template represented by URI and
// locates on the remote system.
type RemoteRegistry struct {
	FetchRetry int
}

func NewRemoteRegistry() *RemoteRegistry {
	return &RemoteRegistry{FetchRetry: 3}
}

func (r *RemoteRegistry) LocateURI(name string) string {
	return name
}

func (r *RemoteRegistry) Find(templateName string) (*RegistryTemplate, error) {
	reader, err := r.fetch(templateName)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	tmpl, err := parseResourceTemplate(reader)
	if err != nil {
		return nil, err
	}
	rt := &RegistryTemplate{
		Name:     templateName,
		Template: tmpl,
		source:   r,
	}
	return rt, nil
}

func (r *RemoteRegistry) LoadRaw(templateName string) ([]byte, error) {
	buf := new(bytes.Buffer)
	reader, err := r.fetch(templateName)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *RemoteRegistry) fetch(templateName string) (io.ReadCloser, error) {
	uri, err := url.Parse(templateName)
	if err != nil {
		return nil, err
	}
	switch uri.Scheme {
	case "http", "https":
		var res *http.Response
		var err error
		for i := 0; i < r.FetchRetry; i++ {
			res, err = http.Get(uri.String())
			if err == nil {
				break
			}
			log.WithError(err).Warnf("http.Get failed retrying... %d/%d", i+1, r.FetchRetry)
		}
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		switch res.StatusCode {
		case http.StatusOK:
			// Pass through.
		case http.StatusNotFound:
			return nil, ErrUnknownTemplateName
		default:
			return nil, fmt.Errorf("Invalid response from remote server: %s", res.Status)
		}

		return res.Body, nil
	default:
		return nil, fmt.Errorf("Unsupported download method: %s", uri.String())
	}
}
