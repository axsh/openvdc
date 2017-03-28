package registry

import (
	"bytes"
	"io"
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
	return "file://" + filepath.ToSlash(abs)
}

func (r *LocalRegistry) Find(templateName string) (*RegistryTemplate, error) {
	f, err := r.open(templateName)
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
		Template: tmpl,
		source:   r,
	}
	return rt, nil
}

func (r *LocalRegistry) LoadRaw(templateName string) ([]byte, error) {
	buf := new(bytes.Buffer)
	f, err := r.open(templateName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := io.Copy(buf, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *LocalRegistry) open(templateName string) (*os.File, error) {
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
	return f, nil
}
