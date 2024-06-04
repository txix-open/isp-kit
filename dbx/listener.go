package dbx

import (
	"context"
	"net"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

type ListenerHandler func(ctx context.Context, msg []byte)

type Listener struct {
	sync.Mutex
	name     string
	doneChan chan struct{}
	fn       ListenerHandler
	logger   log.Logger
}

func (c *Client) NewListener(logger log.Logger, name string, fn ListenerHandler) (*Listener, error) {
	doneChan := make(chan struct{})

	l := &Listener{
		name:     name,
		doneChan: doneChan,
		fn:       fn,
		logger:   logger,
	}

	return l, nil
}

func (c *Client) UpgradeListener(ctx context.Context, l *Listener) error {
	l.Lock()
	defer l.Unlock()

	close(l.doneChan)

	doneChan := make(chan struct{})

	l.doneChan = doneChan

	go c.start(ctx, l.name, l)

	return nil
}

func (l *Listener) Close() error {
	close(l.doneChan)
	return nil
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

		go func(ctx context.Context, connWithWaitNotification *pgx.Conn, l *Listener) {
			select {
			case <-ctx.Done():
				return
			case <-l.doneChan:
				_ = connWithWaitNotification.Close(ctx)
				return
			}
		}(ctx, connWithWaitNotification, l)

		go func(ctx context.Context, connWithWaitNotification *pgx.Conn, l *Listener) {
			for {
				note, err := connWithWaitNotification.WaitForNotification(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) {
						return
					}
					if l.logger != nil {
						l.logger.Warn(ctx, "error listening "+name+": "+err.Error())
					}
					continue
				}
				if len(note.Payload) > 0 {
					nextCtx := ctx
					if len(requestid.FromContext(ctx)) == 0 {
						nextCtx = requestid.ToContext(ctx, requestid.Next())
					}

					l.fn(nextCtx, []byte(note.Payload))
				}
			}
		}(ctx, connWithWaitNotification, l)

		return nil
	})

	if err != nil {
		panic("error connecting to database: " + err.Error())
	}
}
