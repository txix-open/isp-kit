package sentry

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/pkg/errors"
)

const (
	RequestIdKey = "requestId"

	defaultTransportTimeout = 3 * time.Second
	defaultEventBufferSize  = 10
)

type SdkHub struct {
	hub *sentry.Hub
}

func NewHubFromConfiguration(config Config) (Hub, error) {
	if !config.Enable {
		return NewNoopHub(), nil
	}

	if config.Dsn == "" {
		return nil, errors.New("sentry is enabled. dsn must be specified. check sentry configuration")
	}

	buffedTransport := sentry.NewHTTPTransport()
	buffedTransport.Timeout = defaultTransportTimeout
	buffedTransport.BufferSize = defaultEventBufferSize

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         config.Dsn,
		Transport:   buffedTransport,
		ServerName:  config.ModuleName,
		Release:     config.ModuleVersion,
		Environment: config.Environment,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "create sdk client")
	}

	hub := sentry.NewHub(client, sentry.NewScope())

	return SdkHub{
		hub: hub,
	}, nil
}

func (s SdkHub) CatchError(ctx context.Context, err error, level log.Level) {
	eventLevel := sentry.LevelError
	switch level {
	case log.FatalLevel:
		eventLevel = sentry.LevelFatal
	}
	event := &sentry.Event{
		Level:     eventLevel,
		Message:   err.Error(),
		Timestamp: time.Now(),
	}
	event.SetException(err, s.hub.Client().Options().MaxErrorDepth)

	requestId := requestid.FromContext(ctx)
	if requestId != "" {
		event.Extra = map[string]interface{}{
			RequestIdKey: requestId,
		}
	}

	s.CatchEvent(ctx, event)
}

func (s SdkHub) CatchEvent(ctx context.Context, event *sentry.Event) {
	s.hub.CaptureEvent(event)
}

func (s SdkHub) Flush() {
	s.hub.Flush(defaultTransportTimeout)
}
