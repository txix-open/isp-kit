package endpoint

import (
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
)

type JsonResponseMapper struct {
}

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
