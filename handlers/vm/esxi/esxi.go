package esxi

import (
	"encoding/json"
	"flag"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func init() {
	handlers.RegisterHandler(&EsxiHandler{})
}

type EsxiHandler struct {
	vm.Base
}

func (h *EsxiHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.EsxiTemplate{EsxiImage: &model.EsxiTemplate_Image{}}

	type EsxiImage struct {
		Name      string `json:"name,omitempty"`
		Datastore string `json:"datastore,omitempty"`
	}

	var json_template struct {
		EsxiImage EsxiImage `json:"esxi_image,omitempty"`
	}
	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, errors.Wrap(err, "Failed json.Unmarshal EsxiImage")
	}

	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, errors.Wrap(err, "Failed json.Unmarshal for model.EsxiTemplate")
	}

	if tmpl.GetEsxiImage() == nil {
		return nil, handlers.ErrInvalidTemplate(h, "esxi_image must exist")
	}

	return tmpl, nil
}

func (h *EsxiHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	esxiTmpl, ok := m.(*model.EsxiTemplate)
	if !ok {
		panic("template type is not *model.EsxiTemplate")
	}
	t.Item = &model.Template_Esxi{
		Esxi: esxiTmpl,
	}
}

func (h *EsxiHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.EsxiTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.EsxiTemplate), dst)
	}
	msrc, ok := src.(*model.EsxiTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.EsxiTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *EsxiHandler) MergeArgs(dst model.ResourceTemplate, args []string) error {
	mdst, ok := dst.(*model.EsxiTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.EsxiTemplate), dst)
	}

	flags := flag.NewFlagSet("esxi template", flag.ContinueOnError)
	var vcpu, mem int
	flags.IntVar(&vcpu, "vcpu", int(mdst.MinVcpu), "")
	flags.IntVar(&mem, "memory_gb", int(mdst.MinMemoryGb), "")
	if err := flags.Parse(args); err != nil {
		return errors.Wrapf(err, "Failed to parse %v", args)
	}
	mdst.Vcpu = int32(vcpu)
	mdst.MemoryGb = int32(mem)
	return nil
}

func (h *EsxiHandler) Usage(out io.Writer) error {
	return nil
}

func (h *EsxiHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.EsxiTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.EsxiTemplate), dst)
	}
	minput := &model.EsxiTemplate{}
	if err := json.Unmarshal(in, minput); err != nil {
		return errors.WithStack(err)
	}

	proto.Merge(mdst, minput)
	return nil
}
