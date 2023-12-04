package endpoint

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricStorage interface {
	ObserveDuration(method string, duration time.Duration)
	ObserveRequestBodySize(method string, size int)
	ObserveResponseBodySize(method string, size int)
	CountStatusCode(endpoint string, code codes.Code)
}

func Metrics(storage MetricStorage) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			md, _ := metadata.FromIncomingContext(ctx)
			endpoint, err := grpc.StringFromMd(grpc.ProxyMethodNameHeader, md)
			if err != nil {
				return nil, err
			}

			storage.ObserveRequestBodySize(endpoint, len(message.GetBytesBody()))
			start := time.Now()
			response, err := next(ctx, message)
			storage.ObserveDuration(endpoint, time.Since(start))
			storage.CountStatusCode(endpoint, status.Code(err))
			if response != nil {
				storage.ObserveResponseBodySize(endpoint, len(response.GetBytesBody()))
			}

			return response, err
		}
	}
}
