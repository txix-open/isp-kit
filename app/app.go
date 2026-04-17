// Package app provides a lightweight application lifecycle management framework.
//
// The Application struct orchestrates the startup and shutdown of multiple
// Runner components, handling graceful lifecycle transitions and resource cleanup.
//
// # Usage
//
// Create a new application with optional configuration:
//
//	app, err := app.New(
//		app.WithConfigOptions(config.WithExtraSource(config.NewYamlConfig("config.yaml"))),
//		app.WithLoggerConfigSupplier(func(cfg *config.Config) log.Config {
//			return *log.DefaultConfig()
//		}),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Add runners and closers:
//
//	app.AddRunners(myRunner)
//	app.AddClosers(myCloser)
//
// Start the application:
//
//	if err := app.Run(); err != nil {
//		log.Fatal(err)
//	}
//
// Shutdown when needed:
//
//	app.Shutdown()
package app

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/log"
)

// Application manages the lifecycle of long-running services.
// It coordinates the startup and graceful shutdown of multiple Runner components,
// providing context propagation and resource cleanup through Closers.
//
// Application is not safe for concurrent modification. All AddRunners and AddClosers
// calls should be made before calling Run.
type Application struct {
	ctx    context.Context
	cfg    *config.Config
	logger *log.Adapter

	cancel  context.CancelFunc
	runners []Runner
	closers []Closer
}

// New creates a new Application instance with the provided options.
// It applies the options to a DefaultConfig and then creates
// the application from the resulting configuration.
//
// Returns an error if the configuration or logger cannot be initialized.
func New(opts ...Option) (*Application, error) {
	appConfig := DefaultConfig()
	for _, opt := range opts {
		opt(appConfig)
	}
	return NewFromConfig(*appConfig)
}

// NewFromConfig creates a new Application from the provided configuration.
// It initializes the configuration, logger, and context, setting up the
// application with a default closer that syncs the logger on shutdown.
//
// Returns an error if the configuration or logger cannot be initialized.
func NewFromConfig(appConfig Config) (*Application, error) {
	cfg, err := config.New(appConfig.ConfigOptions...)
	if err != nil {
		return nil, errors.WithMessage(err, "create config")
	}

	loggerConfig := appConfig.LoggerConfigSupplier(cfg)
	logger, err := log.NewFromConfig(loggerConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "create logger")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Application{
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
		closers: []Closer{CloserFunc(func() error {
			_ = logger.Sync()
			return nil
		})},
		cancel: cancel,
	}, nil
}

// Context returns the application's context.
// The context is cancelled when Shutdown is called.
func (a *Application) Context() context.Context {
	return a.ctx
}

// Config returns the application's configuration.
func (a *Application) Config() *config.Config {
	return a.cfg
}

// Logger returns the application's logger.
func (a *Application) Logger() *log.Adapter {
	return a.logger
}

// AddRunners appends the provided runners to the application.
// Runners are started as goroutines when Run is called.
//
// AddRunners should be called before Run to ensure proper initialization.
func (a *Application) AddRunners(runners ...Runner) {
	a.runners = append(a.runners, runners...)
}

// AddClosers appends the provided closers to the application.
// Closers are invoked during shutdown to release resources.
//
// AddClosers should be called before Run to ensure proper cleanup.
func (a *Application) AddClosers(closers ...Closer) {
	a.closers = append(a.closers, closers...)
}

// Run starts all registered runners and blocks until one of them returns an error
// or the application context is cancelled.
//
// Each runner executes in its own goroutine. If any runner fails, Run returns
// the first error encountered. If the context is cancelled via Shutdown,
// Run returns nil.
//
// Run is safe to call only once per Application instance.
func (a *Application) Run() error {
	errChan := make(chan error)

	for i := range a.runners {
		go func(index int, runner Runner) {
			err := runner.Run(a.ctx)
			if err != nil {
				select {
				case errChan <- errors.WithMessagef(err, "start runner[%d] -> %T", index, runner):
				default:
					a.logger.Error(a.ctx, errors.WithMessagef(err, "start runner[%d] -> %T", index, runner))
				}
			}
		}(i, a.runners[i])
	}

	select {
	case err := <-errChan:
		return err
	case <-a.ctx.Done():
		return nil
	}
}

// Shutdown initiates graceful shutdown by closing all resources and cancelling
// the application context.
//
// Shutdown should be called to terminate running services and release resources.
func (a *Application) Shutdown() {
	a.Close()
	a.cancel()
}

// Close invokes all registered closers to release resources.
// Errors from individual closers are logged but do not halt the shutdown process.
func (a *Application) Close() {
	for i := range a.closers {
		closer := a.closers[i]
		err := closer.Close()
		if err != nil {
			a.logger.Error(a.ctx, errors.WithMessagef(err, "closers[%d] -> %T", i, closer))
		}
	}
}
