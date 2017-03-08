package model

import (
	"reflect"
	"strings"
)

// ResourceTemplate is a marker interface for all resource template structs.
type ResourceTemplate interface {
	isResourceTemplateKind()
	ResourceName() string
}

func (*NoneTemplate) isResourceTemplateKind() {}
func (*NoneTemplate) ResourceName() string    { return "none" }
func (*LxcTemplate) isResourceTemplateKind()  {}
func (*LxcTemplate) ResourceName() string     { return "vm/lxc" }
func (*NullTemplate) isResourceTemplateKind() {}
func (*NullTemplate) ResourceName() string    { return "vm/null" }

// InstanceResource is a marker interface for instance template structs.
type InstanceResource interface {
	isInstanceResourceKind()
}

func (*LxcTemplate) isInstanceResourceKind()  {}
func (*NullTemplate) isInstanceResourceKind() {}

// ResourceTemplate resolves the assigned object type of
// "Template" OneOf field and cast to ResourceTemplate interface.
// So that you can get the resource name in string.
func GetResourceTemplate(tmpl *Template) ResourceTemplate {
	if tmpl == nil {
		return nil
	}
	v := reflect.ValueOf(tmpl.Item)
	fieldName := strings.TrimPrefix(v.Type().String(), "*model.Template_")
	field := v.Elem().FieldByName(fieldName)
	return field.Interface().(ResourceTemplate)
}
