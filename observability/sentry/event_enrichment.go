package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
)

// contextEnrichmentKey is used as a context key for event enrichment functions.
type contextEnrichmentKey struct{}

// contextKeyValue is the global context key for event enrichment.
//
//go:generate stringer -type=contextEnrichmentKey
var (
	contextKeyValue = contextEnrichmentKey{}
)

// EventEnrichment is a function type that modifies a Sentry event.
// It can be used to add custom tags, context, or other metadata to events.
type EventEnrichment func(event *sentry.Event)

// EnrichEvent adds an event enrichment function to the context.
// The enrichment function will be called for all subsequent log events in this context.
// This allows adding custom metadata to events based on runtime context.
//
// Example:
//
//	ctx := sentry.EnrichEvent(ctx, func(event *sentry.Event) {
//		event.Tags["user-id"] = userID
//		event.Tags["request-type"] = "api"
//	})
func EnrichEvent(ctx context.Context, enrichment EventEnrichment) context.Context {
	return context.WithValue(ctx, contextKeyValue, enrichment)
}
