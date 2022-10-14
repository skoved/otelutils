package otelutils

import (
	"context"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Provides a common interface for *SpanExporterConfigs
type SpanExporterConfig interface {
	newSpanExporter(context.Context) (trace.SpanExporter, error)
}

// ConsoleSpanExporterConfig implements SpanExporterConfig. It is used to create
// an otel exporter that writes to a provided io.Writer. os.Stdout is the
// default writer
type ConsoleSpanExporterConfig struct {
	// Sets the export stream format to use JSON.
	PrettyPrint bool
	// Sets the export stream to include timestamps.
	Timestamps bool
	// Sets the export destination stream.
	Writer *io.Writer
}

// Returns a list of stdouttrace.Option based on the values of the fields in c.
// If c.Writer == nil, os.Stdout is passed to stdouttrace.WithWriter
func (c ConsoleSpanExporterConfig) getOptions() []stdouttrace.Option {
	var options []stdouttrace.Option
	if c.PrettyPrint {
		options = append(options, stdouttrace.WithPrettyPrint())
	}
	if !c.Timestamps {
		options = append(options, stdouttrace.WithoutTimestamps())
	}
	if c.Writer == nil {
		options = append(options, stdouttrace.WithWriter(os.Stdout))
	} else {
		options = append(options, stdouttrace.WithWriter(*c.Writer))
	}
	return options
}

// newConsoleExporter returns an stdouttrace.Exporter. The console exporter
// writes to the location specified by the Writer. This could be a file or
// stdout/stderr.
func (c ConsoleSpanExporterConfig) newSpanExporter(ctx context.Context) (trace.SpanExporter, error) {
	return stdouttrace.New(c.getOptions()...)
}

// OtlpGrpcSpanExporterConfig implements SpanExporterConfig. It is used to
// create an otel exporter that sends ended spans to an OTLP collector using
// gRPC.
type OtlpGrpcSpanExporterConfig struct {
	// a compressor for the gRPC client to use when sending requests. It is the
	// responsibility of the caller to ensure that the compressor has been
	// registered with google.golang.org/grpc/encoding. This can be done by
	// encoding.RegisterCompressor. Some compressors auto-register on import,
	// such as gzip by calling `import _ "google.golang.org/grpc/encoding/gzip"`.
	// This has no effect if GrpcConn is provided.
	Compressor string
	// grpc.DialOptions that will be used when making a connection. The options
	// set here will take precedence over any interal dial options used by
	// `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`. This has
	// no effect if GrpcConn is provided.
	DialOptions []grpc.DialOption
	// Set the target endpoint the exporter will connect to. If unset
	// `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` uses
	// `localhost:4317` as the default. This has no effect if GrpcConn is
	// provided.
	Endpoint string
	// If set, GrpcCon will be the gRPC connection used for all communication.
	// This connection precedence over any other option that relates to
	// establishing or persisting a gRPC connection to a target endpoint. It is
	// the callers responsibility to close the passed connection. This has no
	// effect if GrpcConn is provided.
	GrpcConn *grpc.ClientConn
	// The headers that will be used with each gRPC request.
	Headers map[string]string
	// If true, client transport security for the exporter's gRPC connection is
	// disabled. Just like grpc.WithInsecure() (https://pkg.go.dev/google.golang.org/grpc#WithInsecure)
	// does. If false, client securty is required. This has no effect if GrpcConn
	// is provided.
	Insecure bool
	// Sets the minimum amount of time between connection attempts to the target
	// endpoint. This has no effect if GrpcConn is provided.
	ReconnectionPeriod time.Duration
	// Sets the retry policy for transient retryable errors that may be returned
	// by the target endpoint when exporting a batch of spans. If the target
	// endpoint responds with not only a retryable error, but explicitly returns a
	// backoff time in the response. That time will take precedence over these
	// settings. These settings do not define any network retry strategy. That is
	// entirely handled by the gRPC ClientConn. If empty, the default retry policy
	// will be used. It will retry the export 5 seconds after receiving a
	// retryable error and increase exponentially after each error for no more
	// than a total time of 1 minute. Options in the list are applied to the
	// default retry policy.
	RetryOptions []RetryOption
	// defines the default gRPC service config used. This option has no effect if
	// GrpcConn is provided.
	ServiceConfig string
	// TLS Credentials used when talking to the server. This option has no effect
	// if GrpcConn is provided.
	TlsCredentials *credentials.TransportCredentials
	// Sets the max amount of time a client will attempt to export a batch of
	// spans. This takes precedence over any retry settings defined in
	// RetryConfig. Once the time limit is reached the export is abandoned and
	// the batch of spans is dropped. If unset,
	// `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` sets
	// the default timeout to 10 seconds.
	Timeout time.Duration
}

// Returns a list of otlptracegrpc.Option based on the values of the fields in
// o.
func (o OtlpGrpcSpanExporterConfig) getOptions() []otlptracegrpc.Option {
	var options []otlptracegrpc.Option
	if o.Compressor != "" {
		options = append(options, otlptracegrpc.WithCompressor(o.Compressor))
	}
	if len(o.DialOptions) > 0 {
		options = append(options, otlptracegrpc.WithDialOption(o.DialOptions...))
	}
	if o.Endpoint != "" {
		options = append(options, otlptracegrpc.WithEndpoint(o.Endpoint))
	}
	if o.GrpcConn != nil {
		options = append(options, otlptracegrpc.WithGRPCConn(o.GrpcConn))
	}
	if len(o.Headers) > 0 {
		options = append(options, otlptracegrpc.WithHeaders(o.Headers))
	}
	if o.Insecure {
		options = append(options, otlptracegrpc.WithInsecure())
	}
	if o.ReconnectionPeriod != 0 {
		options = append(options, otlptracegrpc.WithReconnectionPeriod(o.ReconnectionPeriod))
	}
	if len(o.RetryOptions) > 0 {
		retryConfig := otlptracegrpc.RetryConfig{}
		for _, option := range o.RetryOptions {
			option(&retryConfig)
		}
		options = append(options, otlptracegrpc.WithRetry(retryConfig))
	}
	if o.ServiceConfig != "" {
		options = append(options, otlptracegrpc.WithServiceConfig(o.ServiceConfig))
	}
	if o.TlsCredentials != nil {
		options = append(options, otlptracegrpc.WithTLSCredentials(*o.TlsCredentials))
	}
	if o.Timeout != 0 {
		options = append(options, otlptracegrpc.WithTimeout(o.Timeout))
	}
	return options
}

func (o OtlpGrpcSpanExporterConfig) newSpanExporter(ctx context.Context) (trace.SpanExporter, error) {
	return otlptracegrpc.New(ctx, o.getOptions()...)
}

// Retry options are used to configure an otlp exporters RetryConfig
type RetryOption func(*otlptracegrpc.RetryConfig)

// Returns a RetryOption that sets retry.Config.Enabled. Enabled indicates
// whether or not to retry sending batches in case of an export failure.
func RetryEnabled(enabled bool) RetryOption {
	return func(rc *otlptracegrpc.RetryConfig) {
		rc.Enabled = enabled
	}
}

// Returns a RetryOption that sets the of retry.Config.InitialInterval.
// InitialInterval sets the time to wait after the first failure before
// retrying.
func RetryInitialInterval(interval time.Duration) RetryOption {
	return func(rc *otlptracegrpc.RetryConfig) {
		rc.InitialInterval = interval
	}
}

// Returns a RetryOption that sets retry.Confiretry.Config.MaxInterval.
// MaxInterval is upper bound on the backoff interval. Once this value is
// reached, the delay between consecutive retries will always be `MaxInterval`.
func RetryMaxInterval(interval time.Duration) RetryOption {
	return func(rc *otlptracegrpc.RetryConfig) {
		rc.MaxInterval = interval
	}
}

// Returns a RetryOption that sets retry.Config.MaxElapsedTime. MaxElapsedTime
// is the maximum amount of time (including retries) spent trying to send a
// request/batch. Once this value is reached, the data is discarded.
func RetryMaxElapsedTime(interval time.Duration) RetryOption {
	return func(rc *otlptracegrpc.RetryConfig) {
		rc.MaxElapsedTime = interval
	}
}
