package model

import (
	"reflect"
	"strings"

	mesos "github.com/mesos/mesos-go/mesosproto"
)

// ResourceTemplate is a marker interface for all resource template structs.
type ResourceTemplate interface {
	isResourceTemplateKind()
	ResourceName() string
	GetScheduledResource() ScheduleHandler
}

type ScheduleHandler interface {
	Schedule(InstanceResource, map[string]*mesos.Offer) (bool, error) // compare with offer and resrouce request.
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

type NoneSchudler struct {
}

func (n *NoneTemplate) GetScheduledResource() ScheduleHandler {
	return new(NoneSchudler)
}

func (*NoneSchudler) Schedule(ir InstanceResource, offer map[string]*mesos.Offer) (bool, error) {
	return true, nil
}

type NullSchudler struct {
}

func (n *NullTemplate) GetScheduledResource() ScheduleHandler {
	return new(NullSchudler)
}

func (*NullSchudler) Schedule(ir InstanceResource, offer map[string]*mesos.Offer) (bool, error) {
	return true, nil
}

type LxcSchudler struct {
}

func (n *LxcTemplate) GetScheduledResource() ScheduleHandler {
	return new(LxcSchudler)
}

type QemuSchudler struct {
}

func (n *QemuTemplate) GetScheduledResource() ScheduleHandler {
	return new(QemuSchudler)
}

type EsxiSchudler struct {
}

func (n *EsxiTemplate) GetScheduledResource() ScheduleHandler {
	return new(EsxiSchudler)
}

func schedule(ir InstanceResource, offer map[string]*mesos.Offer) (bool, error) {
	offerParams := []string{"vcpu", "mem"}
	reqValues := []int32{
		ir.GetVcpu(),
		ir.GetMemoryGb(),
	}

	for _, offer := range storedOffers {
		var index int
		for i, param := range offerParams {
			index = i
			offerValue := int32(getOfferScalar(offer, param))
			if offerValue < reqValues[i] {
				break
			}
		}
		if index == len(offerParams)-1 {
			return true, nil
		}
	}

	return false, nil
}

func (*LxcSchudler) Schedule(ir InstanceResource, offer map[string]*mesos.Offer) (bool, error) {
	ok, err := schedule(ir, offer)
	return ok, err
}

func (*QemuSchudler) Schedule(ir InstanceResource, offer map[string]*mesos.Offer) (bool, error) {
	ok, err := schedule(ir, offer)
	return ok, err
}

func (*EsxiSchudler) Schedule(ir InstanceResource, offer map[string]*mesos.Offer) (bool, error) {
	ok, err := schedule(ir, offer)
	return ok, err
}

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
