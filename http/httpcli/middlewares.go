package httpcli

import (
	"context"
)

func SetContentLength() Middleware {
	transferEncoding := []string{"identity"}
	return func(next RoundTripper) RoundTripper {
		return RoundTripperFunc(func(ctx context.Context, request *Request) (*Response, error) {
			if request.body == nil { // multipart
				return next.RoundTrip(ctx, request)
			}

			request.Raw.ContentLength = int64(len(request.body))
			request.Raw.TransferEncoding = transferEncoding
			return next.RoundTrip(ctx, request)
		})
	}
}
