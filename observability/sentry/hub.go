// Package sentry provides integration with Sentry for error tracking and event monitoring.
//
// This package wraps the Sentry Go SDK to provide a simple interface for capturing errors
// and events from your application. It includes support for log integration, event enrichment,
// and graceful fallback when Sentry is disabled.
//
// # Basic Usage
//
// Create a hub from configuration:
//
//	hub, err := sentry.NewHubFromConfiguration(sentry.Config{
//		Enable:        true,
//		Dsn:           "your-sentry-dsn",
//		ModuleName:    "my-service",
//		ModuleVersion: "v1.0.0",
//		Environment:   "production",
//	})
//
// Wrap a logger to automatically capture log events:
//
//	logger := sentry.WrapErrorLogger(log.Default(), hub)
//	logger.Error(ctx, "something went wrong", log.String("key", "value"))
//
// # Event Enrichment
//
// You can enrich events with custom context using EnrichEvent:
//
//	ctx := sentry.EnrichEvent(ctx, func(event *sentry.Event) {
//		event.Tags["custom-tag"] = "custom-value"
//	})
package sentry

import (
	"context"
	"maps"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

const (
	// RequestIdKey is the key used to store the request ID in event extra data.
	RequestIdKey = "requestId"

	// defaultTransportTimeout is the timeout for sending events to Sentry.
	defaultTransportTimeout = 3 * time.Second
	// defaultEventBufferSize is the buffer size for the Sentry transport.
	defaultEventBufferSize = 10
)

// SdkHub is the concrete implementation of the Hub interface using the Sentry Go SDK.
type SdkHub struct {
	hub *sentry.Hub
}

// NewHubFromConfiguration creates a new Hub from the provided configuration.
// If Enable is false, it returns a NoopHub that discards all events.
// Returns an error if Dsn is empty when Enable is true.
//
// The created Hub is safe for concurrent use.
func NewHubFromConfiguration(config Config) (Hub, error) {
	if !config.Enable {
		return NewNoopHub(), nil
	}

	if config.Dsn == "" {
		return nil, errors.New("sentry is enabled. dsn must be specified. check sentry configuration")
	}

	allTags := map[string]string{
		"version": config.ModuleVersion,
	}
	if config.InstanceId != "" {
		allTags["instanceId"] = config.InstanceId
	}

	maps.Copy(allTags, config.Tags)

	buffedTransport := sentry.NewHTTPTransport()
	buffedTransport.Timeout = defaultTransportTimeout
	buffedTransport.BufferSize = defaultEventBufferSize

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         config.Dsn,
		Transport:   buffedTransport,
		ServerName:  config.ModuleName,
		Environment: config.Environment,
		Release:     "undefined",
		Tags:        allTags,
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			filtered := make([]sentry.Integration, 0, len(integrations))
			ignoredIntegrations := map[string]bool{
				"Modules":     true,
				"Environment": true,
			}
			for _, integration := range integrations {
				if ignoredIntegrations[integration.Name()] {
					continue
				}
				filtered = append(filtered, integration)
			}
			return filtered
		},
	})
	if err != nil {
		return nil, errors.WithMessage(err, "create sdk client")
	}

	hub := sentry.NewHub(client, sentry.NewScope())

	return SdkHub{
		hub: hub,
	}, nil
}

// CatchError captures an error with the specified log level.
// It maps the log level to a Sentry level and extracts the error stack trace.
// If a request ID is present in the context, it is added to the event.
func (s SdkHub) CatchError(ctx context.Context, err error, level log.Level) {
	eventLevel := sentry.LevelError
	levelFromMapping, ok := logLevelMapping[level]
	if ok {
		eventLevel = levelFromMapping
	}
	event := &sentry.Event{
		Level:     eventLevel,
		Message:   err.Error(),
		Timestamp: time.Now(),
	}
	SetException(event, err)

	requestId := requestid.FromContext(ctx)
	if requestId != "" {
		event.Extra = map[string]any{
			RequestIdKey: requestId,
		}
	}

	s.CatchEvent(ctx, event)
}

// CatchEvent captures a Sentry event directly.
// The event is sent asynchronously to Sentry.
func (s SdkHub) CatchEvent(ctx context.Context, event *sentry.Event) {
	s.hub.CaptureEvent(event)
}

// Flush waits for all buffered events to be sent to Sentry.
// It blocks until the transport timeout or until all events are sent.
func (s SdkHub) Flush() {
	s.hub.Flush(defaultTransportTimeout)
}
