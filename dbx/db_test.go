package dbx_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/test"
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

func TestOpenListener(t *testing.T) {
	tst, require := test.New(t)

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

	// создаем dbrx
	db := dbrx.New()
	err = db.Upgrade(ctx, cfg)
	require.NoError(err)

	// получаем dbx.Client
	db2, err := db.DB()
	require.NoError(err)

	// для dbx.Client создаем listener
	l, err := db2.NewListener(ctx, chanName)
	require.NoError(err)

	// счетчики отправки/получения
	var countGet, countSend int

	// ФП получения данных из listener
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-l.DataChan():
				if len(msg) > 0 {
					require.EqualValues("test message", string(msg))
					countGet++
				}
			case err = <-l.ErrChan():
				tst.Logger().Debug(ctx, "err="+err.Error())
			}
		}
	}()

	// ФП отправки сообщения через notify
	go func() {
		time.Sleep(time.Second / 3)
		for {
			_, err = db2.Exec(ctx, "notify "+chanName+", 'test message'")
			if err != nil {
				return
			}
			countSend++
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second / 10)
				continue
			}
		}
	}()

	// даем поработать
	time.Sleep(3 * time.Second)

	// пересоздаем соедение с БД
	cfg.MaxOpenConn = 12 // нужно, чтобы изменение произошло
	err = db.Upgrade(ctx, cfg)
	require.NoError(err)

	// обновляем listener-а
	db2, err = db.DB()
	require.NoError(err)
	err = db2.UpgradeListener(ctx, l)
	require.NoError(err)

	// даем поработать
	time.Sleep(time.Second)

	// проверяем, что получили все, что отправили
	require.EqualValues(countSend, countGet)

	// закрываем соединения
	err = db.Close()
	require.NoError(err)

	l.ListenerClose()
	time.Sleep(time.Second)
}
