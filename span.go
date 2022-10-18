package otelutils

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	otelTrace "go.opentelemetry.io/otel/trace"
)

const (
	libName = "github.com/skoved/otelutils"
	libVer  = "0.0.1"
)

var tracerOpts = []otelTrace.TracerOption{
	otelTrace.WithInstrumentationVersion(libVer),
	otelTrace.WithSchemaURL(semconv.SchemaURL),
}

// Starts a new span and returns the context of the new span and the span.
// The returned span will be a child span of the span assiciated with parentCtx
func StartSpan(parentCtx context.Context, spanName string, spanOpts ...otelTrace.SpanStartOption) (context.Context, otelTrace.Span) {
	return otel.Tracer(libName, tracerOpts...).Start(parentCtx, spanName, spanOpts...)
}

// Record err as an error as an exception span event on span and set the span
// status of span to Error. Should be called when a function throws an error.
func Error(span *otelTrace.Span, err error) {
	(*span).RecordError(err)
	(*span).SetStatus(codes.Error, err.Error())
}

// Set the status of span to OK. Since description is only added to the span if
// status == codes.Error, description is set to empty string.
func StatusOK(span *otelTrace.Span) {
	(*span).SetStatus(codes.Ok, "")
}
