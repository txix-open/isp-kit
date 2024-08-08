package endpoint

import (
	"context"
	"github.com/txix-open/isp-kit/requestid"
	"net/http"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
)

const (
	RequestIdHeader = "x-request-id"
)

type JsonResponseMapper struct {
}

func (j JsonResponseMapper) Map(ctx context.Context, result any, w http.ResponseWriter) error {
	if result == nil {
		return nil
	}

	requestId := requestid.FromContext(ctx)
	if requestId != "" {
		w.Header().Set(RequestIdHeader, requestId)
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.EncodeInto(w, result)
	if err != nil {
		return errors.WithMessage(err, "marshal json")
	}

	return nil
}
