package httpcli

import (
	"context"
)

// SetContentLength is a middleware that sets the Content-Length header for requests
// with a known body size.
//
// Skips setting Content-Length for multipart requests where the size is unknown.
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
