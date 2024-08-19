package logger

import (
	"context"
	"goplate/config"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

type (
	SLog struct {
		log *slog.Logger
	}
)

var (
	TraceKeyInCtx = new(int)
	rp            = strings.NewReplacer("-", "")
)

const (
	AttrTraceID = "trace_id"
)

func GetTraceID(ctx context.Context) string {
	if ctx == nil || ctx == context.TODO() {
		return rp.Replace(uuid.NewString())
	}

	val := ctx.Value(TraceKeyInCtx)
	if val != nil {
		return val.(string)
	}

	return rp.Replace(uuid.NewString())
}

func SetTraceID(ctx context.Context) context.Context {
	if ctx == nil {
		panic("ctx is nil")
	}

	return context.WithValue(ctx, TraceKeyInCtx, GetTraceID(ctx))
}

func (s *SLog) Debug(msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(context.TODO())).Debug(msg, args...)
}

func (s *SLog) DebugContext(ctx context.Context, msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(ctx)).Debug(msg, args...)
}

func (s *SLog) Error(msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(context.TODO())).Error(msg, args...)
}

func (s *SLog) ErrorContext(ctx context.Context, msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(ctx)).Error(msg, args...)
}

func (s *SLog) Info(msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(context.TODO())).Info(msg, args...)
}

func (s *SLog) InfoContext(ctx context.Context, msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(ctx)).Info(msg, args...)
}

func (s *SLog) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(ctx)).Log(ctx, level, msg, args...)
}

func (s *SLog) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	s.log.With(AttrTraceID, GetTraceID(ctx)).LogAttrs(ctx, level, msg, attrs...)
}

func (s *SLog) Warn(msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(context.TODO())).Warn(msg, args...)
}

func (s *SLog) WarnContext(ctx context.Context, msg string, args ...any) {
	s.log.With(AttrTraceID, GetTraceID(ctx)).Warn(msg, args...)
}

func parseStringLevel(level string) slog.Level {
	switch level {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn
	}
	return slog.LevelError
}

func New(cfg *config.Config) *SLog {
	return &SLog{
		log: slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					AddSource: false,
					Level:     parseStringLevel(cfg.LogLevel),
				},
			),
		),
	}
}
