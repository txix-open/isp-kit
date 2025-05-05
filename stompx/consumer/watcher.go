package consumer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

type Watcher struct {
	config        Config
	close         chan struct{}
	shutdownDone  chan struct{}
	mustReconnect *atomic.Bool
	reportErrOnce *sync.Once
}

func NewWatcher(config Config) *Watcher {
	mustReconnect := &atomic.Bool{}
	mustReconnect.Store(true)

	return &Watcher{
		close:         make(chan struct{}),
		shutdownDone:  make(chan struct{}),
		config:        config,
		mustReconnect: mustReconnect,
		reportErrOnce: &sync.Once{},
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	firstSessionErr := make(chan error, 1)
	go w.run(ctx, firstSessionErr)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-firstSessionErr:
		if err != nil {
			w.mustReconnect.Store(false)
		}
		return err
	}
}

func (w *Watcher) Serve(ctx context.Context) {
	firstSessionErr := make(chan error, 1)
	go w.run(ctx, firstSessionErr)
}

func (w *Watcher) run(ctx context.Context, firstSessionErr chan error) {
	defer func() {
		close(w.shutdownDone)
	}()

	for {
		err := w.runSession(firstSessionErr)
		if err == nil { // normal close
			return
		}

		consumer := &Consumer{
			Config: w.config,
		}
		w.config.Observer.Error(consumer, err)

		if !w.mustReconnect.Load() {
			return // prevent goroutine leak for Run if error occurred
		}

		select {
		case <-ctx.Done():
			return
		case <-w.close:
			return // shutdown called
		case <-time.After(w.config.ReconnectTimeout):
		}
	}
}

// nolint:nonamedreturns
func (w *Watcher) runSession(firstSessionErr chan error) (err error) {
	defer func() {
		if err != nil {
			w.reportFirstSessionError(firstSessionErr, err)
		}
	}()

	c, err := New(w.config)
	if err != nil {
		return errors.WithMessage(err, "new consumer")
	}
	defer c.Close()

	w.reportFirstSessionError(firstSessionErr, nil) // to unblock Run

	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Run()
	}()
	select {
	case err := <-errCh:
		return err
	case <-w.close:
		return nil
	}
}

// Shutdown
// Perform graceful shutdown
func (w *Watcher) Shutdown() {
	close(w.close)
	<-w.shutdownDone
}

func (w *Watcher) reportFirstSessionError(firstSessionErr chan error, err error) {
	w.reportErrOnce.Do(func() {
		firstSessionErr <- err
	})
}
