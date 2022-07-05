package endpoint

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Recovery() Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (msg *isp.Message, err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}
				recovered, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", recovered)
				}
				stack := make([]byte, 4<<10)
				length := runtime.Stack(stack, false)
				err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
			}()
			return next(ctx, message)
		}
	}
}

func ErrorHandler(logger log.Logger) Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			result, err := next(ctx, message)
			if err == nil {
				return result, nil
			}
			logger.Error(ctx, err)
			_, ok := status.FromError(err)
			if ok {
				return result, err
			}
			//hide error details to prevent potential security leaks
			return result, status.Error(codes.Internal, "internal service error")
		}
	}
}

func RequestId() Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, errors.New("metadata is expected in context")
			}
			values := md.Get(grpc.RequestIdHeader)
			requestId := ""
			if len(values) > 0 {
				requestId = values[0]
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))

			return next(ctx, message)
		}
	}
}

func BodyLogger(logger log.Logger) Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			logger.Debug(ctx, "request body", log.ByteString("requestBody", message.GetBytesBody()))
			now := time.Now()

			response, err := next(ctx, message)
			if err == nil {
				logger.Debug(ctx,
					"response body",
					log.ByteString("responseBody", response.GetBytesBody()),
					log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
				)
			}

			return response, err
		}
	}
}
