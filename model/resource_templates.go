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
func (*QemuTemplate) isResourceTemplateKind() {}
func (*QemuTemplate) ResourceName() string    { return "vm/qemu" }
func (*EsxiTemplate) isResourceTemplateKind() {}
func (*EsxiTemplate) ResourceName() string    { return "vm/esxi" }
func (*NullTemplate) isResourceTemplateKind() {}
func (*NullTemplate) ResourceName() string    { return "vm/null" }

// InstanceResource is a marker interface for instance template structs.
type InstanceResource interface {
	ResourceTemplate
	isInstanceResourceKind()
	// protobuf message belongs to InstanceResource should have fields below:
	//  int32 vcpu = xx;
	//  int32 memory_gb = xx;
	//  repeated string node_groups = xx;
	GetVcpu() int32
	GetMemoryGb() int32
	GetNodeGroups() []string
}

func (*LxcTemplate) isInstanceResourceKind()  {}
func (*QemuTemplate) isInstanceResourceKind() {}
func (*NullTemplate) isInstanceResourceKind() {}
func (*EsxiTemplate) isInstanceResourceKind() {}

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

func (t *Template) ResourceTemplate() ResourceTemplate {
	return GetResourceTemplate(t)
}

func IsMatchingNodeGroups(res InstanceResource, offered []string) bool {
	findIndex := func(set []string, group string) int {
		for i, v := range set {
			if v == group {
				return i
			}
		}
		return -1
	}
	for _, reqGroup := range res.GetNodeGroups() {
		if findIndex(offered, reqGroup) < 0 {
			return false
		}
	}
	return true
}

// define openvdc's offer
type VDCOffer struct {
	SlaveID   string
	Resources []Resource
}

type Resource struct {
	Name   string
	Type   valueType
	Scalar float64
	Ranges []valueRange
	Set    []string
	// Disk
}

type valueType int32

const (
	ValueScalar valueType = 0
	ValueRanges valueType = 1
	ValueSet    valueType = 2
	ValueText   valueType = 3
)

type valueRange struct {
	Begin uint64
	End   uint64
}
