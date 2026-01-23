package bootstrap

import (
	"encoding/base64"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/cluster"
)

type DefaultEndpoints struct {
	swagger []byte
}

func NewDefaultEndpoints(path string) (DefaultEndpoints, error) {
	swagger, err := os.ReadFile(path)
	if err != nil {
		return DefaultEndpoints{}, errors.WithMessage(err, "read swagger file")
	}

	return DefaultEndpoints{
		swagger: swagger,
	}, nil
}

func (s DefaultEndpoints) endpointDescriptor(basePath string) []cluster.EndpointDescriptor {
	if len(s.swagger) == 0 {
		return nil
	}

	return []cluster.EndpointDescriptor{{
		Path:             basePath + "/swagger",
		Inner:            false,
		UserAuthRequired: false,
		HttpMethod:       http.MethodGet,
		Handler: func() (string, error) {
			encoded := base64.StdEncoding.EncodeToString(s.swagger)

			return encoded, nil
		},
	}}
}
