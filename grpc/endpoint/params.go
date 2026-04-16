package endpoint

import (
	"context"
	"errors"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"google.golang.org/grpc/metadata"
)

// ContextParam creates a ParamMapper for injecting context.Context into handlers.
// The context is passed through from the incoming gRPC request context.
func ContextParam() ParamMapper {
	return ParamMapper{
		Type: "context.Context",
		Builder: func(ctx context.Context, message *isp.Message) (any, error) {
			return ctx, nil
		},
	}
}

// AuthDataParam creates a ParamMapper for injecting grpc.AuthData into handlers.
// Extracts authentication and authorization information from the incoming request metadata.
// Returns an error if metadata is not present in the context.
func AuthDataParam() ParamMapper {
	return ParamMapper{
		Type: "grpc.AuthData",
		Builder: func(ctx context.Context, message *isp.Message) (any, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, errors.New("metadata is expected in context") // nolint:err113
			}
			return grpc.AuthData(md), nil
		},
	}
}
