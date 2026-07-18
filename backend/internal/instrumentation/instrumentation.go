package instrumentation

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var shutdowns []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var errs []error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			errs = append(errs, shutdowns[i](ctx))
		}
		return errors.Join(errs...)
	}
	bail := func(err error) (func(context.Context) error, error) {
		_ = shutdown(ctx)
		return nil, err
	}
	logShutdown, err := setupLogging(ctx, res, cfg)
	if err != nil {
		return bail(err)
	}
	shutdowns = append(shutdowns, logShutdown)

	tp, err := newTraceProvider(ctx, res, cfg)
	if err != nil {
		return bail(err)
	}

	if tp != nil {
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)
		shutdowns = append(shutdowns, tp.Shutdown)
	}

	metricProvider, err := newMetricProvider(ctx, res, cfg)
	if err != nil {
		return bail(err)
	}
	if metricProvider != nil {
		otel.SetMeterProvider(metricProvider)
		shutdowns = append(shutdowns, metricProvider.Shutdown)
	}
	return shutdown, nil
}
