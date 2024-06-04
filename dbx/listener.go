package dbx

import (
	"context"
	"net"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

type Listener struct {
	name     string
	dataChan chan []byte
	errChan  chan error
	doneChan chan struct{}
}

func (c *Client) NewListener(ctx context.Context, name string) (*Listener, error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	l := &Listener{
		name:     name,
		dataChan: dataChan,
		errChan:  errChan,
		doneChan: doneChan,
	}

	go c.start(ctx, name, l)

	return l, nil
}

func (c *Client) UpgradeListener(ctx context.Context, l *Listener) error {
	close(l.doneChan)

	doneChan := make(chan struct{})

	l.doneChan = doneChan

	go c.start(ctx, l.name, l)

	return nil
}

func (l *Listener) ListenerClose() {
	close(l.doneChan)
}

func (l *Listener) ErrChan() <-chan error {
	return l.errChan
}

func (l *Listener) DataChan() <-chan []byte {
	return l.dataChan
}

func (c *Client) start(ctx context.Context, name string, l *Listener) {
	conn, err := c.Conn(ctx)
	if err != nil {
		panic("error connecting to database: " + err.Error())
	}

	err = conn.Raw(func(driverConn any) error {
		conn2, ok := driverConn.(*stdlib.Conn)
		if !ok {
			panic("error casting driver conn to stdlib")
		}

		connWithWaitNotification := conn2.Conn()
		_, err = connWithWaitNotification.Exec(ctx, "listen "+name)
		if err != nil {
			panic("error listening " + name + ": " + err.Error())
		}

		go func() {
			select {
			case <-ctx.Done():
				return
			case <-l.doneChan:
				_ = connWithWaitNotification.Close(ctx)
				return
			}
		}()

		go func() {
			for {
				note, err := connWithWaitNotification.WaitForNotification(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) {
						return
					}
					l.errChan <- err
					continue
				}
				if len(note.Payload) > 0 {
					l.dataChan <- []byte(note.Payload)
				}
			}
		}()

		return nil
	})

	if err != nil {
		panic("error connecting to database: " + err.Error())
	}
}
