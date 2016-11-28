package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// Handle resource template file locates on the local system.
type LocalRegistry struct {
}

func NewLocalRegistry() *LocalRegistry {
	return &LocalRegistry{}
}

func (r *LocalRegistry) LocateURI(name string) string {
	abs := filepath.Clean(name)
	if !filepath.IsAbs(abs) {
		var err error
		abs, err = filepath.Abs(abs)
		if err != nil {
			return ""
		}
	}
	return fmt.Sprintf("file:///%s", abs)
}

func (r *LocalRegistry) Find(templateName string) (*RegistryTemplate, error) {
	abs := filepath.Clean(templateName)
	if !filepath.IsAbs(abs) {
		var err error
		abs, err = filepath.Abs(abs)
		if err != nil {
			return nil, err
		}
	}
	f, err := os.Open(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrUnknownTemplateName
		}
		return nil, err
	}
	defer f.Close()

	tmpl, err := parseResourceTemplate(f)
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
