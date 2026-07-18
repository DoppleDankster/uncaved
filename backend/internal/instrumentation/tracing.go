package instrumentation

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

func newTraceProvider(
	ctx context.Context,
	res *resource.Resource,
	cfg Config,
) (*sdkTrace.TracerProvider, error) {
	if cfg.OTLPEndpoint == "" {
		return nil, nil
	}

	// Auth is supplied via the standard OTEL_EXPORTER_OTLP_HEADERS env var,
	// which the exporter reads on its own.
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpointURL(cfg.otlpSignalURL("v1/traces")),
	)
	if err != nil {
		return nil, err
	}
	sampler := sdkTrace.ParentBased(sdkTrace.TraceIDRatioBased(cfg.TraceSampling))
	provider := sdkTrace.NewTracerProvider(
		sdkTrace.WithResource(res),
		sdkTrace.WithBatcher(exporter),
		sdkTrace.WithSampler(sampler),
	)
	return provider, nil
}
