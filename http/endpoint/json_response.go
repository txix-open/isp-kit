package endpoint

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/requestid"
)

// JsonResponseMapper maps response objects to JSON format.
// It sets the Content-Type header to application/json and includes the request ID in response headers.
type JsonResponseMapper struct {
}

// Map marshals the result to JSON and writes it to the http.ResponseWriter.
// It skips nil results and includes the request ID in the response headers if available.
func (j JsonResponseMapper) Map(ctx context.Context, result any, w http.ResponseWriter) error {
	reqId := requestid.FromContext(ctx)
	if reqId != "" {
		w.Header().Set(requestid.Header, reqId)
	}

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
