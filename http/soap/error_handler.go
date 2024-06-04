package soap

import (
	"context"
	"net/http"

	http2 "gitlab.txix.ru/isp/isp-kit/http"
	"gitlab.txix.ru/isp/isp-kit/http/endpoint"
	"gitlab.txix.ru/isp/isp-kit/log"
)

func ErrorHandler(logger log.Logger) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			err := next(ctx, w, r)
			if err == nil {
				return nil
			}

			logger.Error(ctx, err)

			httpErr, ok := err.(endpoint.HttpError)
			if ok {
				err = httpErr.WriteError(w)
				return err
			}

			// hide error details to prevent potential security leaks
			err = Fault{
				Code:   "Server",
				String: "internal server error",
			}.WriteError(w)

			return err
		}
	}
}
