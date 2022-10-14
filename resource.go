package otelutils

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// ResourceConfig is used to create an otel resources.Resource. All resources
// will be created with schema URL (currently v1.12.0) and TelemetrySDK added as
// well as attributes defined by fields in the ResourceConfig.
type ResourceConfig struct {
	// Slice of attribute.KeyValue to add to resources created. This field should
	// only be used if attribute(s) need to be added but none of the other fields
	// provide the proper attribute. It is recommended to use the version of
	// `go.opentelemetry.io/otel/semconv` that matches the version of the schema
	// URL used by this package to create the attribute.KeyValue.
	Attributes []attribute.KeyValue
	// Add all Container* attributes to resources created.
	Container bool
	// Add an attribute with the id of the container to resources created
	ContainerID bool
	// TODO
	detectors []resource.Detector
	// Add attributes from the OTEL_RESOURCE_ATTRIBUTES environment variable to
	// the resource created. The value will be interpretted as a list of key/value
	// pairs of taking the form `<key1>=<value1>,<key2>=<value2>,...`
	FromEnv bool
	// Add attributes from the host to the resource created
	Host bool
	// Add all the OS* attributes to the resource created
	OS bool
	// Add an attribute with the OS description to the created resource. The
	// value will be equivalent to the output of `uname -snrvm`
	OSDescription bool
	// Add an attribute with the OS type to the created resource.
	OSType bool
	// Add all Process* attributes to the created resource. WARNING: this option
	// will include ProcessCommandArgs. If the command line arguments contain
	// sensitive information, do not enable this option
	Process bool
	// Add an attribute with all the command line arguments passed to the current
	// process (including the command/executable itself) to the created resource.
	// WARNING: If the command line arguments contain sensitive information, do
	// not enable this option
	ProcessCommandArgs bool
	// Add an attribute with the name of the current process executable to the
	// created resource.
	ProcessExecutableName bool
	// Add an attribute with the full path to the current process executable to
	// the created resource.
	ProcessExecutablePath bool
	// Add an attribute with the username of the user that owns the current
	// process to the created resource
	ProcessOwner bool
	// Add an attribute with the ID of the process to the created resource
	ProcessPID bool
	// Add an attribute with an additional description about the runtime of the
	// current process to the created resource
	ProcessRuntimeDescription bool
	// Add an attribute with the name of the runtime of the current process to
	// the created resource
	ProcessRuntimeName bool
	// Add an attribute with the version of the runtime of the current process to
	// the created resource
	ProcessRuntimeVersion bool
}

// convert the fields r to a slice of resource.Option
func (r ResourceConfig) getOptions() []resource.Option {
	var options []resource.Option
	options = append(options, resource.WithSchemaURL(semconv.SchemaURL))
	options = append(options, resource.WithTelemetrySDK())
	if len(r.Attributes) > 0 {
		options = append(options, resource.WithAttributes(r.Attributes...))
	}
	if r.Container {
		options = append(options, resource.WithContainer())
	}
	if r.ContainerID {
		options = append(options, resource.WithContainerID())
	}
	if len(r.detectors) > 0 {
		options = append(options, resource.WithDetectors(r.detectors...))
	}
	if r.FromEnv {
		options = append(options, resource.WithFromEnv())
	}
	if r.Host {
		options = append(options, resource.WithHost())
	}
	if r.OS {
		options = append(options, resource.WithOS())
	}
	if r.OSDescription {
		options = append(options, resource.WithOSDescription())
	}
	if r.OSType {
		options = append(options, resource.WithOSType())
	}
	if r.Process {
		options = append(options, resource.WithProcess())
	}
	if r.ProcessCommandArgs {
		options = append(options, resource.WithProcessCommandArgs())
	}
	if r.ProcessExecutableName {
		options = append(options, resource.WithProcessExecutableName())
	}
	if r.ProcessExecutablePath {
		options = append(options, resource.WithProcessExecutablePath())
	}
	if r.ProcessOwner {
		options = append(options, resource.WithProcessOwner())
	}
	if r.ProcessPID {
		options = append(options, resource.WithProcessPID())
	}
	if r.ProcessRuntimeDescription {
		options = append(options, resource.WithProcessRuntimeDescription())
	}
	if r.ProcessRuntimeName {
		options = append(options, resource.WithProcessRuntimeName())
	}
	if r.ProcessRuntimeVersion {
		options = append(options, resource.WithProcessRuntimeVersion())
	}
	return options
}

// Creates a new resource based on the options from ResourceConfig r by
// merging the default resource with a new resouce with attributes from fields
// in r. Returns an error if the resource could not be created
func (r ResourceConfig) newResource(ctx context.Context) (*resource.Resource, error) {
	newResource, err := resource.New(ctx, r.getOptions()...)
	if err != nil {
		return nil, err
	}

	return resource.Merge(
		resource.Default(),
		newResource,
	)
}
