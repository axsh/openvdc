package handlers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/axsh/openvdc/model"
)

var (
	resourceHandlers = make(map[string]ResourceHandler)
)

type ResourceHandler interface {
	ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error)
	ShowHelp(out io.Writer) error
}

func RegisterHandler(name string, p ResourceHandler) error {
	if _, exists := resourceHandlers[name]; exists {
		return fmt.Errorf("Duplicated name for resource handler: %s", name)
	}
	resourceHandlers[name] = p
	return nil
}

func FindByType(name string) (p ResourceHandler, ok bool) {
	p, ok = resourceHandlers[name]
	return
}
