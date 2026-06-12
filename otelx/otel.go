package otelx

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const DefaultOTLPCollectorEndpoint = "jaeger.logging.svc.cluster.local:4317"

var initOnce sync.Once

func initOpenTelemetry(ctx context.Context, serviceName string, endpoint string) error {
	needInit := false
	initOnce.Do(func() {
		needInit = true
	})
	if !needInit {
		return nil // already initialized, skip
	}

	if endpoint == "" {
		endpoint = DefaultOTLPCollectorEndpoint
	}

	exporterOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	}

	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		return fmt.Errorf("failed to initialize OTLP exporter: %w", err)
	}

	resource, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize OTel resource: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func BaggageToSpanAttributes(ctx context.Context) []attribute.KeyValue {
	members := baggage.FromContext(ctx).Members()
	attributes := make([]attribute.KeyValue, 0, len(members))
	for _, member := range members {
		attributes = append(attributes, attribute.String(member.Key(), member.Value()))
	}
	return attributes
}
