package endpoint

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/log/logutil"
	sentry2 "github.com/integration-system/isp-kit/observability/sentry"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Recovery() grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (msg *isp.Message, err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
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

func ErrorHandler(logger log.Logger) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			result, err := next(ctx, message)
			if err == nil {
				return result, nil
			}

			logFunc := logutil.LogLevelFuncForError(err, logger)
			logContext := sentry2.EnrichEvent(ctx, func(event *sentry.Event) {
				event.Request = sentryRequest(ctx, message)
			})
			logFunc(logContext, err)

			var grpcErr GrpcError
			if errors.As(err, &grpcErr) {
				return result, grpcErr.GrpcStatusError()
			}

			//deprecated approach
			_, ok := status.FromError(err)
			if ok {
				return result, err
			}

			//hide error details to prevent potential security leaks
			return result, status.Error(codes.Internal, "internal service error")
		}
	}
}

func RequestId() grpc.Middleware {
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

func BodyLogger(logger log.Logger) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			logger.Debug(ctx, "grpc handler: request", log.ByteString("requestBody", message.GetBytesBody()))
			now := time.Now()

			response, err := next(ctx, message)
			if err == nil {
				logger.Debug(ctx,
					"grpc handler: response",
					log.ByteString("responseBody", response.GetBytesBody()),
					log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
				)
			}

			return response, err
		}
	}
}

func sentryRequest(ctx context.Context, msg *isp.Message) *sentry.Request {
	md, _ := metadata.FromIncomingContext(ctx)
	endpoint, _ := grpc.StringFromMd(grpc.ProxyMethodNameHeader, md)
	applicationId, _ := grpc.StringFromMd(grpc.ApplicationIdHeader, md)
	return &sentry.Request{
		URL:    endpoint,
		Method: "POST",
		Headers: map[string]string{
			grpc.ApplicationIdHeader: applicationId,
		},
	}
}
