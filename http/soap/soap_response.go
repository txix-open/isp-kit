package soap

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

type ResponseMapper struct {
}

func (j ResponseMapper) Map(result interface{}, w http.ResponseWriter) error {
	w.Header().Set("content-type", ContentType)

	env := Envelope{Body: Body{Content: result}}
	err := xml.NewEncoder(w).Encode(env)
	if err != nil {
		return errors.WithMessage(err, "xml encode envelope")
	}

	return nil
}
