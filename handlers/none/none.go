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

func (h *NoneHandler) ShowHelp(out io.Writer) error {
	return nil
}
