package otelutils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	otelTrace "go.opentelemetry.io/otel/trace"
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

// initializes the otel configuration for the cli
func OtelInit(name, otelExporter string) error {
	serviceName = name
	var exp sdkTrace.SpanExporter
	resource, err := newResource()

	if err != nil {
		return err
	}

	switch otelExporter {
	case consoleExporter:
		exp, err = newConsoleExporter(os.Stdout)
	case fileExporter:
		f, fErr := os.Create(viper.GetString("OTEL_FILE_SPAN_EXPORTER_NAME"))
		if fErr != nil {
			return fErr
		}
		exp, err = newConsoleExporter(f)
	case otlpExporter:
		exp, err = newOtlpExporter()
	default:
		return fmt.Errorf("Exporter type '%s' is not implemented", otelExporter)
	}

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
func OtelEnd() error {
	if err := tp.ForceFlush(context.Background()); err != nil {
		return err
	}
	return tp.Shutdown(context.Background())
}

// GetTraceParentEnv returns the traceparent from ctx as an envvar
func GetTraceParentEnv(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return fmt.Sprintf("%s=%s", traceParent, carrier.Get(strings.ToLower(traceParent)))
}

// Starts a new span and returns the context of the new span and the span.
// The returned span will be a child span of the span assiciated with parentCtx
func StartSpan(parentCtx context.Context, spanName string, opts ...otelTrace.SpanStartOption) (context.Context, otelTrace.Span) {
	return otel.Tracer(serviceName).Start(parentCtx, spanName, opts...)
}

// adds the error to the passed span. Should be called when a function throws an error
func Error(span *otelTrace.Span, err error) {
	(*span).RecordError(err)
	(*span).SetStatus(codes.Error, err.Error())
}
