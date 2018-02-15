package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/axsh/openvdc/model"
)

var (
	resourceHandlers = make(map[string]ResourceHandler)
)

func ErrInvalidTemplate(h ResourceHandler, msg string) error {
	return fmt.Errorf("Invalid template %s: %s", ResourceName(h), msg)
}

func ErrMergeDstType(expected, dst model.ResourceTemplate) error {
	ev := reflect.ValueOf(expected)
	dv := reflect.ValueOf(dst)
	return fmt.Errorf("Merge failed with destination type: expected type %s but %s", ev.Elem().Type(), dv.Elem().Type())
}

func ErrMergeSrcType(expected, src model.ResourceTemplate) error {
	ev := reflect.ValueOf(expected)
	sv := reflect.ValueOf(src)
	return fmt.Errorf("Merge failed with source type: expected type %s but %s", ev.Elem().Type(), sv.Elem().Type())
}

type ResourceHandler interface {
	ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error)
	// Ugly method... due to the "oneof" protobuf type implementation in Go.
	// https://developers.google.com/protocol-buffers/docs/reference/go-generated#oneof
	SetTemplateItem(t *model.Template, m model.ResourceTemplate)
	IsSupportAPI(m string) bool
}

type InstanceResourceHandler interface {
	ResourceHandler
	GetInstanceSchedulerHandler() InstanceScheduleHandler
}

type CLIHandler interface {
	MergeArgs(src model.ResourceTemplate, args []string) error
	MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error
	Usage(out io.Writer) error
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

type InstanceScheduleHandler interface {
	ScheduleInstance(model.InstanceResource, *model.VDCOffer) (bool, error) // compare with offer and resrouce request.
}
