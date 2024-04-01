package endpoint

import (
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/json"
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
