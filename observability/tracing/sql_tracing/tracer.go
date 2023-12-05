package sql_tracing

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/integration-system/isp-kit/metrics/sql_metrics"
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	QueryParametersKey = attribute.Key("pgx.query.parameters")
)

type Tracer struct {
	tracer trace.Tracer
	config Config
}

func NewTracer(tracer trace.Tracer, config Config) Tracer {
	return Tracer{
		tracer: tracer,
		config: config,
	}
}

func (t Tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}

	label := sql_metrics.OperationLabelFromContext(ctx)
	if label == "" {
		return ctx
	}

	attributes := []attribute.KeyValue{
		tracing.RequestId.String(requestid.FromContext(ctx)),
	}
	if t.config.EnableStatement {
		attributes = append(attributes, semconv.DBStatement(data.SQL))
	}
	if t.config.EnableArgs {
		args := make([]string, 0, len(data.Args))
		for _, arg := range data.Args {
			args = append(args, fmt.Sprintf("%+v", arg))
		}
		attributes = append(attributes, QueryParametersKey.StringSlice(args))
	}

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attributes...),
	}

	spanName := fmt.Sprintf("SQL query %s", label)
	ctx, _ = t.tracer.Start(ctx, spanName, opts...)

	return ctx
}

func (t Tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)

	err := data.Err
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}
