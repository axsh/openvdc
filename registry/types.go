package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/axsh/openvdc/model"
)

var (
	ErrLocalCacheNotReady  = errors.New("Resource template registry has not been cached yet")
	ErrUnknownTemplateName = errors.New("Unknown template name")
)

type TemplateRoot struct {
	Title       string `json:"title`
	Description string `json:"description,omitempty"`

	// "template" block is delayed to parse.
	// https://golang.org/pkg/encoding/json/#RawMessage
	RawTemplate map[string]json.RawMessage `json:"template"`
	Template    model.ResourceTemplate     `json:"-"`
}

type RegistryTemplate struct {
	Name     string
	Template *TemplateRoot
	remote   *GithubRegistry
}

// Returns absolute URI for the original location of the resource template.
func (r *RegistryTemplate) LocationURI() string {
	return r.remote.LocateURI(r.Name)
}

type Registry interface {
	Find(templateName string) (*RegistryTemplate, error)
	ValidateCache() bool
	IsCacheObsolete() (bool, error)
	Fetch() error
}

func parseResourceTemplate(in io.Reader) (*TemplateRoot, error) {
	decoder := json.NewDecoder(in)
	root := &TemplateRoot{}
	err := decoder.Decode(root)
	if err != nil {
		return nil, err
	}
	// Delayed parse for "template" key
	var tmpl model.ResourceTemplate
	for tmplName, raw := range root.RawTemplate {
		tmpl = model.NewTemplateByName(tmplName)
		if tmpl == nil {
			return nil, fmt.Errorf("Unknown template name: %s", tmplName)
		}
		err := json.Unmarshal(raw, tmpl)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse JSON in template key")
		}
		// This is "oneof" type so process only the first item.
		break
	}
	if tmpl == nil {
		return nil, fmt.Errorf("Invalid template definition")
	}
	root.Template = tmpl
	return root, nil
}
