package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
)

// SetupOTelSDK initializes OpenTelemetry tracing.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// Define resource with service metadata.
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("api-server"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("deployment.environment", "production"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Initialize OTLP trace exporter.
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("otel-collector.monitoring.svc.cluster.local:4318"),
		// otlptracehttp.WithEndpoint("otel-collector:4318"), // Collector service in GKE
		otlptracehttp.WithInsecure(), // Use HTTP for simplicity
	)
	if err != nil {
		return nil, err
	}

	// Initialize trace provider.
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces (adjust for production)
	)
	otel.SetTracerProvider(traceProvider)
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)

	// Set up context propagation.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Return shutdown function.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			if shutdownErr := fn(ctx); shutdownErr != nil {
				err = shutdownErr
			}
		}
		return err
	}

	return shutdown, nil
}
