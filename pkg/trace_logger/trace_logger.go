package trace_logger

import (
	"context"

	trace_context "github.com/rzaripov1990/trace_ctx"

	"log/slog"
	"os"
	"strings"
)

var (
	logger *slog.Logger
)

func New(logLevel string, color bool) *slog.Logger {
	logger = slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				AddSource: false,
				Level:     parseStringLevel(logLevel),
			},
		),
	)

	return logger
}

func L(ctx context.Context, val *slog.Logger) *slog.Logger {
	return val.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx))
}

func parseStringLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	}
	return slog.LevelError
}
