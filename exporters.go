package otelutils

import (
	"context"
	"io"
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// newConsoleExporter returns a console exporter. The console exporter writes to the location specified by the Writer.
// This could be a file or stdout/stderr.
func newConsoleExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
}

func newOtlpExporter() (trace.SpanExporter, error) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlptracegrpc.WithInsecure(),
	)
	return otlptrace.New(ctx, client)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-cli"),
			semconv.ServiceVersionKey.String("demo"),
		),
	)
}
