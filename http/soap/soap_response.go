package soap

import (
	"context"
	"encoding/xml"
	"github.com/txix-open/isp-kit/requestid"
	"net/http"

	"github.com/pkg/errors"
)

type ResponseMapper struct {
}

func (j ResponseMapper) Map(ctx context.Context, result any, w http.ResponseWriter) error {
	w.Header().Set("content-type", ContentType)

	reqId := requestid.FromContext(ctx)
	if reqId != "" {
		w.Header().Set(requestid.Header, reqId)
	}

	env := Envelope{Body: Body{Content: result}}
	err := xml.NewEncoder(w).Encode(env)
	if err != nil {
		return errors.WithMessage(err, "xml encode envelope")
	}

	return nil
}
