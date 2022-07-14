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
	ObserveDuration(method string, statusCode codes.Code, duration time.Duration)
	ObserveRequestBodySize(method string, size int)
	ObserveResponseBodySize(method string, size int)
}

func Metrics(storage MetricStorage) Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			md, _ := metadata.FromIncomingContext(ctx)
			method, err := grpc.StringFromMd(grpc.ProxyMethodNameHeader, md)
			if err != nil {
				return nil, err
			}

			storage.ObserveRequestBodySize(method, len(message.GetBytesBody()))
			start := time.Now()
			response, err := next(ctx, message)
			storage.ObserveDuration(method, status.Code(err), time.Since(start))
			if response != nil {
				storage.ObserveResponseBodySize(method, len(response.GetBytesBody()))
			}

			return response, err
		}
	}
}
