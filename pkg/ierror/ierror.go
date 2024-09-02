package ierror

import (
	"context"
	"goplate/pkg/generic"
	"goplate/pkg/trace_logger"
	"sync"
)

var log trace_logger.ITraceLogger

// Initialize only once, in main.go
func New(l trace_logger.ITraceLogger) {
	sync.OnceFunc(
		func() {
			log = l
		},
	)
}

// Check error not nil and log
func Assert(ctx context.Context, err error, exclude ...error) (ok bool) {
	ok = err != nil
	if ok {
		if !generic.In(err, exclude) {
			log.ErrorContext(ctx, "", err)
		} else {
			ok = false
		}
	}
	return
}
