package endpoint

import (
	"context"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/codec"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/logutil"
	sentry2 "github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/panic_recovery"
	"github.com/txix-open/isp-kit/requestid"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultEncodeThreshold = codec.DefaultEncodeThreshold
)

// Recovery creates a middleware that catches panics and converts them to errors.
// Prevents gRPC server crashes from handler panics by recovering and returning the error.
func Recovery() grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (msg *isp.Message, err error) {
			defer panic_recovery.Recover(func(panicErr error) {
				err = panicErr
			})
			return next(ctx, message)
		}
	}
}

// ErrorHandler creates a middleware that handles errors from downstream handlers.
// Logs errors at appropriate levels, enriches them with Sentry context, and
// converts custom GrpcError types to gRPC status errors.
// Returns a generic "internal service error" for unknown error types to prevent
// information leakage.
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

// RequestId creates a middleware that manages request IDs for tracing.
// Extracts the request ID from incoming metadata, generates a new one if absent,
// and injects it into the context for downstream use.
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

type DecodeCodec interface {
	Type() string
	DecodeBytes(b []byte) ([]byte, error)
}

// DecodeRequest returns middleware that transparently handles decoded requests.
//
//	If x-content-encoding matches codec, request body is decoded.
//
// Important:
//   - Should be placed before logging middleware if logs must see decoded data.
func DecodeRequest(codec DecodeCodec) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, msg *isp.Message) (*isp.Message, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, errors.New("metadata is expected in context")
			}

			contentEncoding, _ := grpc.StringFromMd(grpc.ContentEncodingHeader, md)
			reqBody := msg.GetBytesBody()
			if contentEncoding == codec.Type() {
				decoded, err := codec.DecodeBytes(reqBody)
				if err != nil {
					return nil, err
				}

				msg.Body = &isp.Message_BytesBody{
					BytesBody: decoded,
				}
				md.Delete(grpc.ContentEncodingHeader)
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			return next(ctx, msg)
		}
	}
}

type EncodeCodec interface {
	Type() string
	EncodeBytes(b []byte) ([]byte, error)
}

// EncodeResponse returns middleware that transparently handles response encoding.
//
// If client supports x-accept-encoding and response body size > encodeThresholdBytes,
// the response is encoded using codec.
//
// Important:
//   - Should be placed before logging middleware if logs must see plain data.
func EncodeResponse(codec EncodeCodec, encodeThresholdBytes int) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, msg *isp.Message) (*isp.Message, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, errors.New("metadata is expected in context")
			}

			resp, err := next(ctx, msg)
			if err != nil {
				return nil, err
			}

			// check accept
			acceptEncoding, _ := grpc.StringFromMd(grpc.AcceptEncodingHeader, md)
			if !acceptsEncoding(acceptEncoding, codec.Type()) {
				return resp, nil
			}

			responseBody := resp.GetBytesBody()
			if len(responseBody) <= encodeThresholdBytes {
				return resp, nil
			}

			err = grpc2.SetHeader(ctx, metadata.Pairs(
				grpc.ContentEncodingHeader, codec.Type(),
			))
			if err != nil {
				return nil, err
			}

			encoded, err := codec.EncodeBytes(responseBody)
			if err != nil {
				return nil, err
			}
			resp.Body = &isp.Message_BytesBody{
				BytesBody: encoded,
			}

			return resp, nil
		}
	}
}

// acceptsEncoding split header and check contains encoding in it.
func acceptsEncoding(header, encoding string) bool {
	for part := range strings.SplitSeq(header, ",") {
		if strings.TrimSpace(part) == encoding {
			return true
		}
	}

	return false
}

// sentryRequest creates a Sentry request object from the gRPC context.
// Extracts endpoint and application ID from metadata for error tracking.
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
