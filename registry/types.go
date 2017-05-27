package registry

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
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

func (t *TemplateRoot) parseTemplate() error {
	// Delayed parse for "template" key
	typeFind := struct {
		Type string `json:"type"`
	}{}
	if err := json.Unmarshal(t.RawTemplate, &typeFind); err != nil {
		return err
	}
	hnd, ok := handlers.FindByType(typeFind.Type)
	if !ok {
		return fmt.Errorf("Unknown template type: %s", typeFind.Type)
	}
	t.handler = hnd

	var err error
	t.Template, err = hnd.ParseTemplate(t.RawTemplate)
	if err != nil {
		return errors.Wrapf(err, "%T found parse error", hnd)
	}

	return nil
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

func (r *RegistryTemplate) ToModel() *model.Template {
	t := &model.Template{
		TemplateUri: r.LocationURI(),
	}
	r.Template.handler.SetTemplateItem(t, r.Template.Template)
	return t
}

type TemplateFinder interface {
	Find(templateName string) (*RegistryTemplate, error)
	LocateURI(templateName string) string
	LoadRaw(templateName string) ([]byte, error)
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

	err = root.parseTemplate()
	if err != nil {
		return nil, err
	}
	return root, nil
}
