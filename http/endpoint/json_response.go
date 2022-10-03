package endpoint

import (
	"net/http"

	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
)

type JsonResponseMapper struct {
}

func (j JsonResponseMapper) Map(result interface{}, w http.ResponseWriter) error {
	if result == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.EncodeInto(w, result)
	if err != nil {
		return errors.WithMessage(err, "marshal json")
	}

	return nil
}
