package none

import (
	"encoding/json"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
)

func init() {
	handlers.RegisterHandler(&NoneHandler{})
}

type NoneHandler struct {
}

func (h *NoneHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.NoneTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (h *NoneHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_None{
		None: m.(*model.NoneTemplate),
	}
}

func (h *NoneHandler) MergeArgs(dst model.ResourceTemplate, args []string) error {
	_, ok := dst.(*model.NoneTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.NoneTemplate), dst)
	}
	return nil
}

func (h *NoneHandler) Usage(out io.Writer) error {
	return nil
}

func (h *NoneHandler) IsSupportAPI(method string) bool {
	return false
}

func (h *NoneHandler) Merge(dst, src model.ResourceTemplate) error {
	if _, ok := dst.(*model.NoneTemplate); !ok {
		return handlers.ErrMergeDstType(new(model.NoneTemplate), dst)
	}
	if _, ok := src.(*model.NoneTemplate); !ok {
		return handlers.ErrMergeSrcType(new(model.NoneTemplate), src)
	}
	return nil
}
