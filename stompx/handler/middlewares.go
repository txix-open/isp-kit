package handler

import (
	"context"
	"runtime"

	"github.com/go-stomp/stomp/v3"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
)

const (
	panicStackLength = 4 << 10
)

type Middleware func(next HandlerAdapter) HandlerAdapter

func Log(logger log.Logger) Middleware {
	return func(next HandlerAdapter) HandlerAdapter {
		return AdapterFunc(func(ctx context.Context, msg *stomp.Message) Result {
			destination := msg.Destination

			result := next.Handle(ctx, msg)

			switch {
			case result.Ack:
				logger.Debug(
					ctx,
					"stomp client: message will be acknowledged",
					log.String("destination", destination),
				)
			case result.Requeue:
				logger.Error(
					ctx,
					"stomp client: message will be requeued",
					log.Any("error", result.Err),
					log.String("destination", destination),
				)
			}

			return result
		})
	}
}

// nolint:nonamedreturns
func Recovery() Middleware {
	return func(next HandlerAdapter) HandlerAdapter {
		return AdapterFunc(func(ctx context.Context, msg *stomp.Message) (res Result) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				var err error
				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					err = errors.Errorf("%v", recovered)
				}
				stack := make([]byte, panicStackLength)
				length := runtime.Stack(stack, false)
				res.Err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
				res.Requeue = true
			}()
			return next.Handle(ctx, msg)
		})
	}
}
