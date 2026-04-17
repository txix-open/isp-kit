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
	"github.com/txix-open/isp-kit/requestid"
)

// DBProvider defines an interface for obtaining a database client.
type DBProvider interface {
	DB() (*dbx.Client, error)
}

// Client manages background job workers and provides functionality for
// enqueuing and processing jobs. It handles worker lifecycle, including
// startup, graceful shutdown, and configuration updates.
//
// The Client is safe for concurrent use by multiple goroutines.
type Client struct {
	db      DBProvider
	logger  log.Logger
	lock    sync.Locker
	workers []*bgjob.Worker
}

// NewClient creates a new Client instance with the provided database provider
// and logger. The client is ready to be configured with workers via Upgrade.
func NewClient(db DBProvider, logger log.Logger) *Client {
	return &Client{
		db:     db,
		logger: logger,
		lock:   &sync.Mutex{},
	}
}

// Upgrade configures and starts workers based on the provided worker configurations.
// It gracefully stops any previously running workers before starting the new ones.
// Each worker polls its designated queue for jobs and processes them using the
// configured handler.
//
// Returns an error if the database connection cannot be established or if the
// job store cannot be created.
func (c *Client) Upgrade(ctx context.Context, workerConfigs []WorkerConfig) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.shutdownAllWorkers()

	db, err := c.db.DB()
	if err != nil {
		return errors.WithMessage(err, "get db")
	}
	store, err := bgjob.NewPgStoreV2(ctx, db.DB.DB)
	if err != nil {
		return errors.WithMessage(err, "create bgjob store")
	}
	cli := bgjob.NewClient(store)

	workers := make([]*bgjob.Worker, 0)
	metricStorage := bgjob_metrics.NewStorage(metrics.DefaultRegistry)
	for _, config := range workerConfigs {
		worker := bgjob.NewWorker(
			cli,
			config.Queue,
			NewDefaultHandler(config.Handle, metricStorage),
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

// Enqueue adds a single job to the background job queue.
// If the request does not contain a RequestId, it attempts to extract one from
// the context, or generates a new one if none is present.
//
// Returns an error if the database connection cannot be established.
func (c *Client) Enqueue(ctx context.Context, req bgjob.EnqueueRequest) error {
	db, err := c.db.DB()
	if err != nil {
		return errors.WithMessage(err, "get db")
	}

	requestId := req.RequestId

	if requestId == "" {
		requestId = requestid.FromContext(ctx)
	}
	if requestId == "" {
		requestId = requestid.Next()
	}
	req.RequestId = requestId

	return bgjob.Enqueue(ctx, db, req)
}

// BulkEnqueue adds multiple jobs to the background job queue in a single operation.
// If any request in the list does not contain a RequestId, it inherits the main
// RequestId from the context (or generates one if not present).
//
// Returns an error if the database connection cannot be established.
func (c *Client) BulkEnqueue(ctx context.Context, list []bgjob.EnqueueRequest) error {
	db, err := c.db.DB()
	if err != nil {
		return errors.WithMessage(err, "get db")
	}

	mainRequestId := requestid.FromContext(ctx)
	if mainRequestId == "" {
		mainRequestId = requestid.Next()
	}

	for i := range list {
		if list[i].RequestId == "" {
			list[i].RequestId = mainRequestId
		}
	}
	return bgjob.BulkEnqueue(ctx, db, list)
}

// Close gracefully shuts down all running workers.
// This method should be called when the application is stopping to ensure
// in-flight jobs are properly handled.
func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.shutdownAllWorkers()
}

// shutdownAllWorkers stops all registered workers by calling their Shutdown method.
// This method assumes the caller holds the lock.
func (c *Client) shutdownAllWorkers() {
	for _, worker := range c.workers {
		worker.Shutdown()
	}
}
