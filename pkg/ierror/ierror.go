package ierror

import (
	"context"
	"log/slog"
	"sync"

	generic "github.com/rzaripov1990/genx"
)

var logger *slog.Logger

// Initialize only once, in main.go
func New(l *slog.Logger) {
	sync.OnceFunc(
		func() {
			logger = l
		},
	)
}

// Check error not nil and log
func Assert(ctx context.Context, err error, exclude ...error) (ok bool) {
	ok = err != nil
	if ok {
		if !generic.In(err, exclude) {
			logger.ErrorContext(ctx, "", "received error", err.Error())
		} else {
			ok = false
		}
	}
	return
}
