package traceid

import (
	"context"
	"log/slog"
)

const LogKey = "trace_id"

type LogHandler struct {
	slog.Handler
	LogKey string
}

func NewLogHandler(next slog.Handler) *LogHandler {
	return &LogHandler{
		Handler: next,
		LogKey:  LogKey,
	}
}

func (h *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := FromContext(ctx); ok {
		r = r.Clone()
		r.AddAttrs(slog.String(h.LogKey, id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLogHandler(h.Handler.WithAttrs(attrs))
}

func (h *LogHandler) WithGroup(name string) slog.Handler {
	return NewLogHandler(h.Handler.WithGroup(name))
}
