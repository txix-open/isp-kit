package stompx

import (
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/handler"
)

// NewResultHandler creates a result handler with logging and panic recovery.
func NewResultHandler(logger log.Logger, adapter handler.HandlerAdapter) handler.ResultHandler {
	return handler.NewHandler(
		logger,
		adapter,
		handler.Log(logger),
		handler.Recovery(),
	)
}
