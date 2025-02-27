package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type OTELOption func(*otelOptions)

type otelOptions struct {
	endpoint string
	interval time.Duration
}

func WithOTELEndpoint(endpoint string) OTELOption {
	return func(o *otelOptions) {
		o.endpoint = endpoint
	}
}

func WithOTELInterval(interval time.Duration) OTELOption {
	return func(o *otelOptions) {
		o.interval = interval
	}
}

// StartOTEL bootstraps the OpenTelemetry pipeline. The collected metrics are
// exported to the specified OTLP (opentelemetry protocol) collector HTTP
// endpoint. The endpoint is in host port format, with no schema, as it uses
// unencrypted HTTP currently. If it does not return an error, make sure to call
// shutdown for proper cleanup.
func StartOTEL(ctx context.Context, options ...OTELOption) (func(context.Context) error, error) {
	opts := &otelOptions{
		endpoint: "127.0.0.1:4318",
		interval: 10 * time.Second,
	}
	for _, o := range options {
		o(opts)
	}

	// TODO: maybe permit a full url and use WithEndpointURL, and we append the
	// required path like /v1/traces or /v1/metrics, which is currently done
	// automatically when using WithEndpoint(hostport).

	// Define the resource attributes (e.g., service name)
	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String("kwil-db")), // i.e. service.name = kwil-db, but with a fancy package
		resource.WithOS(), resource.WithProcess(), resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) error {
		return errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	// prop := newPropagator()
	// otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	traceExporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(opts.endpoint),
		otlptracehttp.WithInsecure(),
	)
	// traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, handleErr(err)
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(opts.interval)),
	)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	otel.SetTracerProvider(tracerProvider) // for use with otel.Tracer()

	// Set up meter exporter.
	metricExporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpoint(opts.endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, handleErr(err)
	}

	// Set up meter provider.
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(opts.interval))), // Default is 1m.
	)

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	// loggerProvider, err := newLoggerProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return
	// }
	// shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	// global.SetLoggerProvider(loggerProvider)

	return shutdown, nil
}

// func newPropagator() propagation.TextMapPropagator {
// 	return propagation.NewCompositeTextMapPropagator(
// 		propagation.TraceContext{},
// 		propagation.Baggage{},
// 	)
// }

// func newLoggerProvider() (*log.LoggerProvider, error) {
// 	logExporter, err := stdoutlog.New()
// 	if err != nil {
// 		return nil, err
// 	}

// 	loggerProvider := log.NewLoggerProvider(
// 		log.WithProcessor(log.NewBatchProcessor(logExporter)),
// 	)
// 	return loggerProvider, nil
// }
