package endpoint

import (
	"context"
	"fmt"
	"runtime"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/logutil"
	sentry2 "github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/requestid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	panicStackLength = 4 << 10
)

// nolint:nonamedreturns
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
					err = fmt.Errorf("%v", recovered) // nolint:err113,errorlint
				}
				stack := make([]byte, panicStackLength)
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
				event.Request = sentryRequest(ctx)
			})
			logFunc(logContext, err)

			var grpcErr GrpcError
			if errors.As(err, &grpcErr) {
				return result, grpcErr.GrpcStatusError()
			}

			// deprecated approach
			_, ok := status.FromError(err)
			if ok {
				return result, err
			}

			// hide error details to prevent potential security leaks
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
			values := md.Get(requestid.Header)
			requestId := ""
			if len(values) > 0 {
				requestId = values[0]
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String(requestid.LogKey, requestId))

			return next(ctx, message)
		}
	}
}

func sentryRequest(ctx context.Context) *sentry.Request {
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
