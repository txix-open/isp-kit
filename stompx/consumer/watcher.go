package consumer

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type Watcher struct {
	config       Config
	close        chan struct{}
	shutdownDone chan struct{}
}

func NewWatcher(config Config) *Watcher {
	return &Watcher{
		close:        make(chan struct{}),
		shutdownDone: make(chan struct{}),
		config:       config,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	firstSessionErr := make(chan error, 1)
	go w.run(ctx, firstSessionErr)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-firstSessionErr:
		return err
	}
}

func (w *Watcher) run(ctx context.Context, firstSessionErr chan error) {
	defer func() {
		close(w.shutdownDone)
	}()

	sessNum := 0
	for {
		sessNum++
		err := w.runSession(sessNum, firstSessionErr)
		if err == nil { //normal close
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-w.close:
			return //shutdown called
		case <-time.After(w.config.ReconnectTimeout):

		}
	}
}

func (w *Watcher) runSession(sessNum int, firstSessionErr chan error) (err error) {
	firstSessionErrWritten := false
	defer func() {
		if !firstSessionErrWritten && sessNum == 1 {
			firstSessionErr <- err
		}
	}()

	c, err := New(w.config)
	if err != nil {
		return errors.WithMessage(err, "new consumer")
	}
	defer c.Close()

	if sessNum == 1 {
		firstSessionErrWritten = true
		firstSessionErr <- nil
	}

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
