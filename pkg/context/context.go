package context

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

var (
	TraceKeyInCtx any = new(byte)
	rp                = strings.NewReplacer("-", "")
)

const (
	TraceIDKeyName = "trace_id"
)

func genTraceID() string {
	return rp.Replace(uuid.NewString())
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return genTraceID()
	}

	val := ctx.Value(TraceKeyInCtx)
	if val != nil {
		return val.(string)
	}

	return genTraceID()
}

func WithTraceID(ctx context.Context) context.Context {
	if ctx == nil {
		panic("ctx is nil")
	}

	return context.WithValue(ctx, TraceKeyInCtx, genTraceID())
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		panic("ctx is nil")
	}

	return context.WithValue(ctx, TraceKeyInCtx, traceID)
}
