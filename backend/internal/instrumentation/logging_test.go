package instrumentation

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

// twoJSONHandlers builds a fanout over two independent buffers so a test can
// assert both destinations saw the same record.
func twoJSONHandlers(level slog.Level) (fanout, *bytes.Buffer, *bytes.Buffer) {
	var a, b bytes.Buffer
	opts := &slog.HandlerOptions{Level: level}
	return fanout{
		level:    level,
		handlers: []slog.Handler{slog.NewJSONHandler(&a, opts), slog.NewJSONHandler(&b, opts)},
	}, &a, &b
}

func TestFanoutHandleReachesEveryChild(t *testing.T) {
	f, a, b := twoJSONHandlers(slog.LevelInfo)
	log := slog.New(f)

	log.Info("hello", "k", "v")

	for name, buf := range map[string]*bytes.Buffer{"a": a, "b": b} {
		if !strings.Contains(buf.String(), `"msg":"hello"`) || !strings.Contains(buf.String(), `"k":"v"`) {
			t.Fatalf("handler %s missing record: %q", name, buf.String())
		}
	}
}

func TestFanoutEnabledGate(t *testing.T) {
	f, a, b := twoJSONHandlers(slog.LevelWarn)
	log := slog.New(f)

	log.Info("dropped")

	if a.Len() != 0 || b.Len() != 0 {
		t.Fatalf("info leaked past warn gate: a=%q b=%q", a.String(), b.String())
	}
}

// TestFanoutWithAttrsSizesByHandlers guards the WithAttrs slice-length bug:
// a single attr against two handlers must not panic and must reach both.
func TestFanoutWithAttrsSizesByHandlers(t *testing.T) {
	f, a, b := twoJSONHandlers(slog.LevelInfo)
	log := slog.New(f).With("request_id", "abc")

	log.Info("with attrs")

	for name, buf := range map[string]*bytes.Buffer{"a": a, "b": b} {
		if !strings.Contains(buf.String(), `"request_id":"abc"`) {
			t.Fatalf("handler %s missing With attr: %q", name, buf.String())
		}
	}
}

func TestNewHandlerStdoutOnlyWhenNoProvider(t *testing.T) {
	h := newHandler(Config{}, slog.LevelInfo, nil)
	f, ok := h.(fanout)
	if !ok {
		t.Fatalf("expected fanout, got %T", h)
	}
	if len(f.handlers) != 1 {
		t.Fatalf("expected stdout-only (1 handler), got %d", len(f.handlers))
	}
}

func TestNewStdoutHandlerFormatByEnvironment(t *testing.T) {
	if _, ok := newStdoutHandler(Config{Environment: "development"}, slog.LevelInfo).(*slog.TextHandler); !ok {
		t.Fatalf("development should use TextHandler")
	}
	if _, ok := newStdoutHandler(Config{Environment: "production"}, slog.LevelInfo).(*slog.JSONHandler); !ok {
		t.Fatalf("non-development should use JSONHandler")
	}
}

func TestSetupLoggingOffDiscards(t *testing.T) {
	shutdown, err := setupLogging(context.Background(), nil, Config{LogLevel: "off"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DiscardHandler.Enabled is false for every level; that is what "off" means.
	if slog.Default().Enabled(context.Background(), slog.LevelError) {
		t.Fatalf("off should install a discarding logger")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("off shutdown should be a noop, got %v", err)
	}
}

func TestSetupLoggingStdoutOnly(t *testing.T) {
	shutdown, err := setupLogging(context.Background(), nil, Config{LogLevel: "debug"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected a shutdown func")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("stdout-only shutdown should be a noop, got %v", err)
	}
}

// TestFanoutHandleRace fans concurrent records across two handlers; without
// r.Clone() the shared record's attr backing array races. Run with -race.
func TestFanoutHandleRace(t *testing.T) {
	f, _, _ := twoJSONHandlers(slog.LevelInfo)
	log := slog.New(f)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.LogAttrs(context.Background(), slog.LevelInfo, "concurrent",
				slog.Int("i", i), slog.String("s", "x"))
		}(i)
	}
	wg.Wait()
}
