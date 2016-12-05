package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/axsh/openvdc/handlers"
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
	RawTemplate json.RawMessage        `json:"template"`
	Template    model.ResourceTemplate `json:"-"`

	handler handlers.ResourceHandler
}

func (t *TemplateRoot) ResourceHandler() handlers.ResourceHandler {
	return t.handler
}

type RegistryTemplate struct {
	Name     string
	Template *TemplateRoot
	source   TemplateFinder
}

// Returns absolute URI for the original location of the resource template.
func (r *RegistryTemplate) LocationURI() string {
	return r.source.LocateURI(r.Name)
}

type TemplateFinder interface {
	Find(templateName string) (*RegistryTemplate, error)
	LocateURI(templateName string) string
}

type CachedRegistry interface {
	TemplateFinder
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
	typeFind := struct {
		Type string `json:"type"`
	}{}
	if err := json.Unmarshal(root.RawTemplate, &typeFind); err != nil {
		return nil, err
	}
	hnd, ok := handlers.FindByType(typeFind.Type)
	if !ok {
		return nil, fmt.Errorf("Unknown template type: %s", typeFind.Type)
	}
	tmpl, err := hnd.ParseTemplate(root.RawTemplate)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse template section in %s", root.Title)
	}
	root.Template = tmpl
	root.handler = hnd
	return root, nil
}
