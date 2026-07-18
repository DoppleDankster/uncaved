package instrumentation

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkMetrics "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func newMetricProvider(
	ctx context.Context,
	res *resource.Resource,
	cfg Config,
) (*sdkMetrics.MeterProvider, error) {
	if cfg.OTLPEndpoint == "" {
		return nil, nil
	}
	exporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpointURL(cfg.otlpSignalURL("v1/metrics")),
	)
	if err != nil {
		return nil, err
	}
	reader := sdkMetrics.NewPeriodicReader(exporter, sdkMetrics.WithInterval(cfg.MetricInterval))

	return sdkMetrics.NewMeterProvider(
		sdkMetrics.WithResource(res),
		sdkMetrics.WithReader(reader),
	), nil
}
