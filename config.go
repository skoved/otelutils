package otelutils

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// supported otel exporters
const (
	fileExporter    = "file"
	consoleExporter = "console"
	otlpExporter    = "otlp"
)

// Name of the trace parent envvar
const traceParent = "TRACEPARENT"

var (
	tp          *sdkTrace.TracerProvider
	serviceName string
)

// initializes the otel configuration for the cli.
func OtelInit(startCtx context.Context, resourceConf ResourceConfig, exporterConfig SpanExporterConfig) error {
	resource, err := resourceConf.newResource(startCtx)
	if err != nil {
		return err
	}

	exp, err := exporterConfig.newSpanExporter(startCtx)
	if err != nil {
		return err
	}

	tp = sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(exp),
		sdkTrace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

// flush and shutdown the global TracerProvider
func OtelEnd(endCtx context.Context) error {
	if err := tp.ForceFlush(endCtx); err != nil {
		return err
	}
	return tp.Shutdown(endCtx)
}

// GetTraceParentEnv returns the traceparent from ctx as an envvar
func GetTraceParentEnv(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return fmt.Sprintf("%s=%s", traceParent, carrier.Get(strings.ToLower(traceParent)))
}
