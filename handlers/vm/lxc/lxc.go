package lxc

import (
	"encoding/json"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
)

func init() {
	handlers.RegisterHandler(&LxcHandler{})
}

type LxcHandler struct {
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.LxcTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (h *LxcHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Lxc{
		Lxc: m.(*model.LxcTemplate),
	}
}

func (h *LxcHandler) Usage(out io.Writer) error {
	return nil
}
