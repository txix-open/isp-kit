package soap

import (
	"context"
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/requestid"
)

// ResponseMapper maps response objects to SOAP XML format.
// It wraps the result in a SOAP envelope and sets the appropriate content type.
type ResponseMapper struct {
}

// Map encodes the result as a SOAP envelope and writes it to the http.ResponseWriter.
// It sets the Content-Type to text/xml and includes the request ID in response headers if available.
func (j ResponseMapper) Map(ctx context.Context, result any, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", ContentType)

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
