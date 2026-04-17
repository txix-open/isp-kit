package endpoint

import (
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/json"
)

// JsonResponseMapper maps handler results to gRPC messages with JSON-encoded bodies.
// Implements the ResponseBodyMapper interface.
type JsonResponseMapper struct{}

// Map marshals the result to JSON and creates a gRPC message.
// Returns an empty message body if result is nil.
// Returns an error if JSON marshaling fails.
func (j JsonResponseMapper) Map(result any) (*isp.Message, error) {
	if result == nil {
		return &isp.Message{Body: &isp.Message_BytesBody{}}, nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal json")
	}
	return &isp.Message{
		Body: &isp.Message_BytesBody{
			BytesBody: data,
		},
	}, nil
}
