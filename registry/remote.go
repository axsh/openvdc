package registry

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Handle resource template represented by URI and
// locates on the remote system.
type RemoteRegistry struct {
}

func NewRemoteRegistry() *RemoteRegistry {
	return &RemoteRegistry{}
}

func (r *RemoteRegistry) LocateURI(name string) string {
	return name
}

func (r *RemoteRegistry) Find(templateName string) (*RegistryTemplate, error) {
	uri, err := url.Parse(templateName)
	if err != nil {
		return nil, err
	}

	var input io.Reader
	switch uri.Scheme {
	case "http", "https":
		res, err := http.Get(uri.String())
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

		input = res.Body
	default:
		return nil, fmt.Errorf("Unsupported download method: %s", uri.String())
	}
	tmpl, err := parseResourceTemplate(input)
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
