package tracing

import (
	"context"

	"github.com/go-logr/stdr"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// DefaultPropagator is the default text map propagator using W3C Trace Context.
var DefaultPropagator Propagator = propagation.TraceContext{}

// DefaultProvider is the global default tracer provider, initially set to a no-op provider.
var DefaultProvider TracerProvider = NewNoopProvider()

// RequestId is the attribute key used to store the request ID in spans.
const (
	RequestId = attribute.Key("app.request_id")
)

// NewProviderFromConfiguration creates a new tracer provider from the given configuration.
// It returns a no-op provider if tracing is disabled. The provider is configured to export
// traces via OTLP over HTTP to the specified address.
func NewProviderFromConfiguration(ctx context.Context, logger log.Logger, config Config) (Provider, error) {
	if !config.Enable {
		return NewNoopProvider(), nil
	}

	stdLogger := log.StdLoggerWithLevel(logger, log.InfoLevel, log.String("worker", "tracer"))
	otel.SetLogger(stdr.New(stdLogger))

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(config.Address),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "new otlp http exporter")
	}

	attributes := []attribute.KeyValue{
		semconv.DeploymentEnvironment(config.Environment),
		semconv.ServiceVersion(config.ModuleVersion),
		semconv.ServiceName(config.ModuleName),
		semconv.ServiceInstanceID(config.InstanceId),
	}
	for key, value := range config.Attributes {
		attributes = append(attributes, attribute.String(key, value))
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			attributes...,
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "new resource")
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return provider, nil
}
