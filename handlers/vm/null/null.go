package null

import (
	"encoding/json"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
)

func init() {
	handlers.RegisterHandler(&NullHandler{})
}

type NullHandler struct {
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

func (h *NullHandler) Usage(out io.Writer) error {
	return nil
}
