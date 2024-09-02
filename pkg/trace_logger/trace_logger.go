// Package `trace_logger` provides a wrapper for the standard `slog` package,
// adding enhanced logging capabilities.
//
// Specifically, it includes the use of a unique log entry identifier, `trace_id`,
// within functions, which allows for efficient tracking and correlation
// of all log entries related to a single request.
//
// This significantly simplifies diagnostics and process analysis,
// providing a deeper understanding of what is happening within the system.
package trace_logger

import (
	"context"
	"goplate/env"
	"goplate/pkg/trace_context"
	"log/slog"
	"os"
	"strings"
)

type (
	ITraceLogger interface {
		Debug(msg string, args ...any)
		DebugContext(ctx context.Context, msg string, args ...any)
		Error(msg string, args ...any)
		ErrorContext(ctx context.Context, msg string, args ...any)
		Info(msg string, args ...any)
		InfoContext(ctx context.Context, msg string, args ...any)
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
		LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
		Warn(msg string, args ...any)
		WarnContext(ctx context.Context, msg string, args ...any)
	}

	TraceLoggerJson struct {
		l *slog.Logger
	}
)

func New(cfg *env.BaseConfig) ITraceLogger {
	return &TraceLoggerJson{
		l: slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					AddSource: false,
					Level:     parseStringLevel(cfg.Log.Level),
				},
			),
		),
	}
}

func (s *TraceLoggerJson) Debug(msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(context.TODO())).Debug(msg, args...)
}

func (s *TraceLoggerJson) DebugContext(ctx context.Context, msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx)).Debug(msg, args...)
}

func (s *TraceLoggerJson) Error(msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(context.TODO())).Error(msg, args...)
}

func (s *TraceLoggerJson) ErrorContext(ctx context.Context, msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx)).Error(msg, args...)
}

func (s *TraceLoggerJson) Info(msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(context.TODO())).Info(msg, args...)
}

func (s *TraceLoggerJson) InfoContext(ctx context.Context, msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx)).Info(msg, args...)
}

func (s *TraceLoggerJson) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx)).Log(ctx, level, msg, args...)
}

func (s *TraceLoggerJson) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx)).LogAttrs(ctx, level, msg, attrs...)
}

func (s *TraceLoggerJson) Warn(msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(context.TODO())).Warn(msg, args...)
}

func (s *TraceLoggerJson) WarnContext(ctx context.Context, msg string, args ...any) {
	s.l.With(trace_context.TraceIDKeyName, trace_context.GetTraceID(ctx)).Warn(msg, args...)
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
