package endpoint

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
)

type JsonResponseMapper struct {
}

func (j JsonResponseMapper) Map(result any, w http.ResponseWriter) error {
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
