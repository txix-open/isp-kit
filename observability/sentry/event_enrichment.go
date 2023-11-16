package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
)

type contextEnrichmentKey struct{}

var (
	contextKeyValue = contextEnrichmentKey{}
)

type EventEnrichment func(event *sentry.Event)

func LazilyEnrichEvent(ctx context.Context, enrichment EventEnrichment) context.Context {
	return context.WithValue(ctx, contextKeyValue, enrichment)
}
