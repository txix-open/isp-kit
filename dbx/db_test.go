package dbx_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/dbx"
)

func TestOpen(t *testing.T) {
	require := require.New(t)
	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := dbx.Config{
		Host:     envOrDefault("PG_HOST", "127.0.0.1"),
		Port:     port,
		Database: envOrDefault("PG_DB", "test"),
		Username: envOrDefault("PG_USER", "test"),
		Password: envOrDefault("PG_PASS", "test"),
	}
	db, err := dbx.Open(context.Background(), cfg)
	require.NoError(err)
	var time time.Time
	err = db.SelectRow(context.Background(), &time, "select now()")
	require.NoError(err)
}

func envOrDefault(name string, defValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}
	return defValue
}

func TestOpenNativeListener(t *testing.T) {
	require := require.New(t)
	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := dbx.Config{
		Host:     envOrDefault("PG_HOST", "127.0.0.1"),
		Port:     port,
		Database: envOrDefault("PG_DB", "test"),
		Username: envOrDefault("PG_USER", "test"),
		Password: envOrDefault("PG_PASS", "test"),
	}

	ctx, cancel := context.WithCancel(context.Background())

	db, err := dbx.Open(ctx, cfg)
	require.NoError(err)

	conn, err := db.Conn(ctx)
	require.NoError(err)

	var count1, count2 int

	err = conn.Raw(func(driverConn any) error {
		conn2, ok := driverConn.(*stdlib.Conn)
		require.EqualValues(true, ok)

		// connWithWait := conn2.Conn()
		_, err = conn2.Conn().Exec(ctx, "listen channelname")
		require.NoError(err)

		go func() {
			for {
				n, err := conn2.Conn().WaitForNotification(ctx)
				if err != nil {
					// println("error occurred waiting for notification")
					if errors.Is(err, context.Canceled) {
						println("finish wait")
						return
					}
					fmt.Printf("err=%#v\n", err)
					continue
				}
				count1++
				println("n=" + n.Payload)
			}
		}()

		return nil
	})
	require.NoError(err)

	go func() {
		time.Sleep(time.Second / 3)
		for {
			_, err = db.Exec(ctx, "SELECT pg_notify('channelname', 'test_message')")
			if err != nil && errors.Is(err, context.Canceled) {
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				count2++
				continue
			}
		}
	}()

	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(time.Second)
	require.EqualValues(count1, count2)
}

func TestOpenListener(t *testing.T) {
	require := require.New(t)
	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := dbx.Config{
		Host:        envOrDefault("PG_HOST", "127.0.0.1"),
		Port:        port,
		Database:    envOrDefault("PG_DB", "test"),
		Username:    envOrDefault("PG_USER", "test"),
		Password:    envOrDefault("PG_PASS", "test"),
		MaxOpenConn: 10,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chanName := "tmpChan"

	db := dbrx.New()
	err = db.Upgrade(ctx, cfg)
	require.NoError(err)

	db2, err := db.DB()
	require.NoError(err)

	dataChan, errChan, err := db2.NewListener(ctx, chanName)
	require.NoError(err)

	var count1, count2 int

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-dataChan:
				require.EqualValues("test message", string(msg))
				count1++
			case err = <-errChan:
				println("err=" + err.Error())
			}
		}
	}()

	go func() {
		time.Sleep(time.Second / 3)
		for {
			_, err = db2.Exec(ctx, "notify "+chanName+", 'test message'")
			if err != nil {
				fmt.Printf("error for sender::%#v\n", err)
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				count2++
				time.Sleep(time.Second / 10)
				continue
			}
		}
	}()

	time.Sleep(3 * time.Second)
	// cancel()
	println("call upgrade::1::", count1, count2)
	cfg.MaxOpenConn = 12
	err = db.Upgrade(ctx, cfg)
	require.NoError(err)
	println("call upgrade::2::", count1, count2)
	time.Sleep(time.Second)
	println("call upgrade::3::", count1, count2)
	require.EqualValues(count2, count1)
}
