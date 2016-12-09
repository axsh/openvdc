package registry

import (
	"fmt"
	"io"

	"github.com/xeipuuv/gojsonschema"
)

type ErrInvalidTemplate struct {
	Errors []gojsonschema.ResultError
}

func (e *ErrInvalidTemplate) Error() string {
	return fmt.Sprintf("Detected %d errors", len(e.Errors))
}

func (e *ErrInvalidTemplate) Dump(out io.Writer) {
	for _, r := range e.Errors {
		out.Write([]byte(r.Description() + ": " + r.Field() + "\n"))
	}
}

func ValidateTemplate(in []byte) error {
	buf, err := Asset("schema/v1.json")
	if err != nil {
		return err
	}
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(buf))
	if err != nil {
		return err
	}
	r, err := schema.Validate(gojsonschema.NewBytesLoader(in))
	if err != nil {
		return err
	}
	if !r.Valid() {
		return &ErrInvalidTemplate{Errors: r.Errors()}
	}
	return nil
}
