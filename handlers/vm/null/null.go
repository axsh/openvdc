package null

import (
	"encoding/json"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/gogo/protobuf/proto"
)

func init() {
	handlers.RegisterHandler(&NullHandler{})
}

type NullHandler struct {
	vm.Base
}

func (h *NullHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.NullTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (h *NullHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Null{
		Null: m.(*model.NullTemplate),
	}
}

func (h *NullHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.NullTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.NullTemplate), dst)
	}
	msrc, ok := src.(*model.NullTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.NullTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *NullHandler) Usage(out io.Writer) error {
	return nil
}
