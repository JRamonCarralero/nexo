// Package logger builds the application's structured logger (log/slog).
//
// In production it emits JSON using the field names Google Cloud Logging
// understands (severity, message). Every record logged with a context that
// carries a trace ID includes it automatically.
package logger

import (
	"context"
	"io"
	"log/slog"
)

type traceIDKey struct{}

// WithTraceID returns a context carrying the request's trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// TraceID returns the trace ID stored in ctx, or "".
func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey{}).(string)
	return id
}

// New builds a logger that writes to w. Development gets readable text; any other env gets JSON.
func New(w io.Writer, env string, level slog.Level) *slog.Logger {
	var h slog.Handler
	if env == "development" {
		h = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	} else {
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			AddSource:   false,
			ReplaceAttr: cloudLoggingAttrs,
		})
	}
	return slog.New(contextHandler{h})
}

// cloudLoggingAttrs renames the standard keys to the ones Cloud Logging expects.
func cloudLoggingAttrs(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}
	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
		if lvl, ok := a.Value.Any().(slog.Level); ok && lvl == slog.LevelWarn {
			a.Value = slog.StringValue("WARNING")
		}
	case slog.MessageKey:
		a.Key = "message"
	}
	return a
}

// contextHandler adds the trace ID from the context to every record.
type contextHandler struct{ slog.Handler }

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := TraceID(ctx); id != "" {
		r.AddAttrs(slog.String("trace_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}
