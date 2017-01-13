package model

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"golang.org/x/net/context"
)

type ResourceOps interface {
	Create(*Resource) (*Resource, error)
	FindByID(string) (*Resource, error)
	Destroy(string) error
}

const resourcesBaseKey = "resources"

func init() {
	schemaKeys = append(schemaKeys, resourcesBaseKey)
}

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

// ResourceTemplate resolves the assigned object type of
// "Template" OneOf field and cast to ResourceTemplate interface.
// So that you can get the resource name in string.
func (r *Resource) ResourceTemplate() ResourceTemplate {
	if r.Template == nil {
		return nil
	}
	v := reflect.ValueOf(r.Template.Item)
	fieldName := strings.TrimPrefix(v.Type().String(), "*model.Template_")
	field := v.Elem().FieldByName(fieldName)
	return field.Interface().(ResourceTemplate)
}

type resources struct {
	base
}

func Resources(ctx context.Context) ResourceOps {
	return &resources{base{ctx: ctx}}
}

func (i *resources) Create(n *Resource) (*Resource, error) {
	n.State = Resource_REGISTERED
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	nkey, err := bk.CreateWithID(fmt.Sprintf("/%s/r-", resourcesBaseKey), n)
	if err != nil {
		return nil, err
	}
	n.Id = path.Base(nkey)
	return n, nil
}

func (i *resources) FindByID(id string) (*Resource, error) {
	bk, err := i.connection()
	if err != nil {
		return nil, err
	}
	n := &Resource{}
	if err := bk.Find(fmt.Sprintf("/%s/%s", resourcesBaseKey, id), n); err != nil {
		return nil, err
	}
	n.Id = id
	return n, nil
}

func (i *resources) Destroy(id string) error {
	bk, err := i.connection()
	if err != nil {
		return err
	}
	n, err := i.FindByID(id)
	if err != nil {
		return err
	}
	if err := n.validateStateTransition(Resource_UNREGISTERED); err != nil {
		return err
	}
	n.State = Resource_UNREGISTERED
	return bk.Update(fmt.Sprintf("/%s/%s", resourcesBaseKey, id), n)
}

func (r *Resource) validateStateTransition(next Resource_State) error {
	var result bool
	switch r.GetState() {
	case Resource_REGISTERED:
		result = (next == Resource_UNREGISTERED)
	}

	if result {
		return nil
	}

	return fmt.Errorf("Invalid state transition: %s -> %s",
		r.GetState().String(),
		next.String(),
	)
}
