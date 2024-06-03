package dbx

import (
	"context"
	"net"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

// TODO: завтра
//		новый объект Listener (data, error, done)
//		у объекта Close, которая закрывает done
//		когда done закрыт, то мы закрываем connWith...

func (c *Client) ListenerClose() {

}

func (c *Client) NewListener(ctx context.Context, name string) (chan []byte, chan error, error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)

	go c.start(ctx, dataChan, errChan, name)

	return dataChan, errChan, nil
}

func (c *Client) start(ctx context.Context, dataChan chan []byte, errChan chan error, name string) {
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
			println("старт wait")
			defer println("выход из wait")
			for {
				note, err := connWithWaitNotification.WaitForNotification(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, net.ErrClosed) {
						return
					}
					errChan <- err
					continue
				}
				dataChan <- []byte(note.Payload)
			}
		}()

		return nil
	})

	if err != nil {
		panic("error connecting to database: " + err.Error())
	}
}
