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

//go:generate go-bindata-assetfs -mode 420 -pkg registry -prefix ../ ../schema/...

func ValidateTemplate(in []byte) error {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoaderFileSystem("file://schema/v1.json", assetFS()))
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
