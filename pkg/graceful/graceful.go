// The `graceful` package allows for the proper management of resource startup and shutdown,
// ensuring that resources are initialized and terminated correctly to maintain system stability and reliability.
package graceful

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

type (
	Notify struct {
		Signals []os.Signal
	}

	OnError struct {
		Handler any
		Func    func(ctx context.Context, handler any, err error)
	}

	CloseGroup struct {
		shutdownCtx context.Context //nolint:containedctx
		cancel      context.CancelFunc

		closer []closer
		errors chan error

		onError OnError

		wait atomic.Int32
	}

	closer struct {
		resource any
		close    any
		task     task
	}

	Task[T_resource any] func(ctx context.Context, resource T_resource) error

	task func(ctx context.Context, resource, close any, errors chan error)
)

var (
	DefaultNotify = Notify{
		Signals: []os.Signal{os.Interrupt},
	}

	DefaultOnError = OnError{
		Handler: log.Default(),
		Func:    defaultOnError,
	}
)

func defaultOnError(_ context.Context, handler any, err error) {
	handler.(*log.Logger).Println(err)
}

func Prepare(ctx context.Context, options ...any) (shutdownCtx context.Context, group *CloseGroup) {
	// make group and set defaults
	group = &CloseGroup{
		errors:  make(chan error),
		onError: DefaultOnError,
	}

	notify := DefaultNotify

	// read all known options, ignore unknown
	for _, option := range options {
		switch typed := option.(type) {
		case Notify:
			notify = typed
		case OnError:
			group.onError = typed
		}
	}

	if len(notify.Signals) > 0 {
		// listen os signals
		group.shutdownCtx, group.cancel = signal.NotifyContext(ctx, notify.Signals...)
	} else {
		group.shutdownCtx, group.cancel = context.WithCancel(ctx)
	}

	return group.shutdownCtx, group
}

func Process[T_resource any](group *CloseGroup, resource T_resource, process Task[T_resource]) {
	group.wait.Add(1)
	go wrapProcess(group, resource, process)
}

func Close[T_resource any](group *CloseGroup, resource T_resource, taskClose Task[T_resource]) {
	// fill closer with data before running goroutine to avoid errors
	group.closer = append(group.closer, closer{
		resource: resource,
		close:    taskClose,
		task: func(ctx context.Context, data, close any, errors chan error) {
			// we need data and close as any to pass it obviously
			// types for cast will be inherited from parent function at compile time
			errors <- close.(Task[T_resource])(ctx, data.(T_resource))
		},
	})
}

func wrapProcess[T_resource any](group *CloseGroup, resource T_resource, process Task[T_resource]) {
	err := process(group.shutdownCtx, resource)

	// cancel shutdown context if error acquired
	group.cancel()
	// wait until main goroutine receive error
	group.errors <- err
}

func wrapClose[T_resource any](ctx context.Context, data, taskClose any, errors chan error) {
	errors <- taskClose.(Task[T_resource])(ctx, data.(T_resource))
}

func (group *CloseGroup) Wait(timeout time.Duration) {
	// wait until shutdown context cancelled
	<-group.shutdownCtx.Done()

	// shutdown context cancelled there so we need in separate context for closing
	ctx := context.Background()

	if timeout > 0 {
		// some closers may take time to close, it may be illegal for os so better set context with timeout
		withTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		ctx = withTimeout
	}

	// run closers as goroutines in parallel
	for _, closer := range group.closer {
		group.wait.Add(1)
		go closer.task(ctx, closer.resource, closer.close, group.errors)
	}

cycle:
	for group.wait.Add(-1) >= 0 {
		select {
		case err := <-group.errors:
			if err != nil {
				group.onError.Func(ctx, group.onError.Handler, err)
			}
		case <-ctx.Done():
			// finish if context cancelled faster than all errors processed
			break cycle
		}
	}
}
