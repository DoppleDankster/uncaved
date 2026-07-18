package instrumentation

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

const scopeName = "github.com/DoppleDankster/uncaved"

const defaultLogLevel = slog.LevelInfo

// parseLevel returns a slog.Level from a string using UnmarshalText
// default value in case of an empty string is slog.LevelInfo
func parseLevel(s string) (slog.Level, error) {
	if s == "" {
		return defaultLogLevel, nil
	}
	var lvl slog.Level
	err := lvl.UnmarshalText([]byte(s))
	return lvl, err
}

// newLoggerProvider returns a configured logger provider
// configured for OTLP HTTP with batch processing
func newLoggerProvider(
	ctx context.Context,
	res *resource.Resource,
	cfg Config,
) (*sdklog.LoggerProvider, error) {
	if cfg.OTLPEndpoint == "" {
		return nil, nil
	}

	exporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithEndpointURL(cfg.otlpSignalURL("v1/logs")),
	)
	if err != nil {
		return nil, err
	}
	bp := sdklog.NewBatchProcessor(exporter)
	provider := sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(bp))
	return provider, nil
}

// Fanout Handler to output logs to multiple downstream handlers
// implements the slog.Handler interface
type fanout struct {
	level    slog.Level
	handlers []slog.Handler
}

func (f fanout) Enabled(_ context.Context, l slog.Level) bool {
	return l >= f.level.Level()
}

func (f fanout) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range f.handlers {
		if !handler.Enabled(ctx, r.Level) {
			continue
		}
		if err := handler.Handle(ctx, r.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (f fanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(f.handlers))
	for i, handler := range f.handlers {
		next[i] = handler.WithAttrs(attrs)
	}
	return &fanout{level: f.level, handlers: next}
}

func (f fanout) WithGroup(name string) slog.Handler {
	if name == "" {
		return f
	}
	next := make([]slog.Handler, len(f.handlers))
	for i, handler := range f.handlers {
		next[i] = handler.WithGroup(name)
	}
	return &fanout{level: f.level, handlers: next}
}

// traceContext is a slog.Handler middleware that stamps the active span's
// trace_id and span_id onto each record, so stdout logs can be correlated with
// traces. The OTLP bridge (otelslog) derives these from the context itself;
// this covers the plain text/JSON stdout handler, which is span-unaware.
// Records logged without a span-carrying context (i.e. via slog.Info rather
// than slog.InfoContext(c.Request.Context(), ...)) pass through untouched.
type traceContext struct {
	slog.Handler
}

func (t traceContext) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r = r.Clone()
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return t.Handler.Handle(ctx, r)
}

func (t traceContext) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceContext{t.Handler.WithAttrs(attrs)}
}

func (t traceContext) WithGroup(name string) slog.Handler {
	return traceContext{t.Handler.WithGroup(name)}
}

// newStdoutHandler builds the terminal handler: human-readable text in
// development, JSON everywhere else so a collector can parse it.
func newStdoutHandler(cfg Config, lvl slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: lvl}
	if cfg.Environment == "development" {
		return slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.NewJSONHandler(os.Stdout, opts)
}

// newHandler assembles the slog handler the application logs through: stdout
// always, plus the OTLP bridge when a provider is present.
// lvl gates both destinations uniformly, since otelslog delegates Enabled to the provider,
// which accepts everything.
func newHandler(cfg Config, lvl slog.Level, lp *sdklog.LoggerProvider) slog.Handler {
	// Wrap the stdout handler so console logs carry trace_id/span_id; the
	// otelslog bridge below already correlates on its own, so it stays bare.
	handlers := []slog.Handler{traceContext{newStdoutHandler(cfg, lvl)}}
	if lp != nil {
		handlers = append(handlers,
			otelslog.NewHandler(scopeName, otelslog.WithLoggerProvider(lp)))
	}
	return fanout{level: lvl, handlers: handlers}
}

// setupLogging installs the process-wide slog logger and returns a shutdown
// func that flushes the OTLP batch processor.
func setupLogging(
	ctx context.Context,
	res *resource.Resource,
	cfg Config,
) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	// "off" disables logging entirely
	if cfg.LogLevel == "off" {
		slog.SetDefault(slog.New(slog.DiscardHandler))
		return noop, nil
	}

	lvl, err := parseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	lp, err := newLoggerProvider(ctx, res, cfg)
	if err != nil {
		return nil, err
	}

	slog.SetDefault(slog.New(newHandler(cfg, lvl, lp)))

	// Route OTel's own diagnostics to a stdout-only handler.
	// Sending them through the default logger would feed export errors back into the failing OTLP path.
	// These are process-global across all three signals.
	diag := slog.New(newStdoutHandler(cfg, lvl))
	otel.SetLogger(logr.FromSlogHandler(diag.Handler()))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		diag.Error("otel", "error", err)
	}))

	if lp == nil {
		return noop, nil
	}
	otellogglobal.SetLoggerProvider(lp)
	return lp.Shutdown, nil
}
