package bgjobx

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/bgjob_metrics"
)

type DBProvider interface {
	DB() (*dbx.Client, error)
}

type Client struct {
	db      DBProvider
	logger  log.Logger
	lock    sync.Locker
	workers []*bgjob.Worker
}

func NewClient(db DBProvider, logger log.Logger) *Client {
	return &Client{
		db:     db,
		logger: logger,
		lock:   &sync.Mutex{},
	}
}

func (c *Client) Upgrade(ctx context.Context, workerConfigs []WorkerConfig) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.shutdownAllWorkers()

	db, err := c.db.DB()
	if err != nil {
		return errors.WithMessage(err, "get db")
	}
	cli := bgjob.NewClient(bgjob.NewPgStore(db.DB.DB))

	workers := make([]*bgjob.Worker, 0)
	metricStorage := bgjob_metrics.NewStorage(metrics.DefaultRegistry)
	for _, config := range workerConfigs {
		worker := bgjob.NewWorker(
			cli,
			config.Queue,
			WithDurationMeasure(metricStorage, config.Handle),
			bgjob.WithConcurrency(config.GetConcurrency()),
			bgjob.WithPollInterval(config.GetPollInterval()),
			bgjob.WithObserver(Observer{log: c.logger, metricStorage: metricStorage}),
		)
		workers = append(workers, worker)
	}
	c.workers = workers

	for _, worker := range c.workers {
		worker.Run(ctx)
	}

	return nil
}

func (c *Client) Enqueue(ctx context.Context, req bgjob.EnqueueRequest) error {
	db, err := c.db.DB()
	if err != nil {
		return errors.WithMessage(err, "get db")
	}
	return bgjob.Enqueue(ctx, db, req)
}

func (c *Client) BulkEnqueue(ctx context.Context, list []bgjob.EnqueueRequest) error {
	db, err := c.db.DB()
	if err != nil {
		return errors.WithMessage(err, "get db")
	}
	return bgjob.BulkEnqueue(ctx, db, list)
}

func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.shutdownAllWorkers()
}

func (c *Client) shutdownAllWorkers() {
	for _, worker := range c.workers {
		worker.Shutdown()
	}
}
