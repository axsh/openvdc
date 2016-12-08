package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"reflect"

	"github.com/axsh/openvdc/model"
)

var (
	resourceHandlers = make(map[string]ResourceHandler)
)

type ResourceHandler interface {
	ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error)
	ShowHelp(out io.Writer) error
	// Ugly method... due to the "oneof" protobuf type implementation in Go.
	// https://developers.google.com/protocol-buffers/docs/reference/go-generated#oneof
	SetTemplateItem(t *model.Template, m model.ResourceTemplate)
}

func RegisterHandler(p ResourceHandler) error {
	name := ResourceName(p)
	if _, exists := resourceHandlers[name]; exists {
		return fmt.Errorf("Duplicated name for resource handler: %s", name)
	}
	resourceHandlers[name] = p
	return nil
}

// ResourceName estimate the resource type name from package path string.
func ResourceName(h ResourceHandler) string {
	v := reflect.ValueOf(h)
	return strings.TrimPrefix(v.Elem().Type().PkgPath(), "github.com/axsh/openvdc/handlers/")
}

func FindByType(name string) (p ResourceHandler, ok bool) {
	p, ok = resourceHandlers[name]
	return
}
