package endpoint

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetricStorage defines the interface for collecting gRPC server metrics.
// Implementations should record request durations, payload sizes, and response codes.
type MetricStorage interface {
	// ObserveDuration records the duration of a request for the given method.
	ObserveDuration(method string, duration time.Duration)
	// ObserveRequestBodySize records the size of the request body.
	ObserveRequestBodySize(method string, size int)
	// ObserveResponseBodySize records the size of the response body.
	ObserveResponseBodySize(method string, size int)
	// CountStatusCode increments the counter for the given endpoint and status code.
	CountStatusCode(endpoint string, code codes.Code)
}

// Metrics creates a middleware that collects metrics for gRPC server requests.
// Records request duration, request/response body sizes, and response status codes.
// The endpoint name is extracted from the ProxyMethodNameHeader in request metadata.
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
