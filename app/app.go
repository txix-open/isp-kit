package app

import (
	"context"

	"github.com/integration-system/isp-kit/config"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
)

type Application struct {
	ctx    context.Context
	cfg    *config.Config
	logger *log.Adapter

	cancel  context.CancelFunc
	runners []Runner
	closers []Closer
}

func New(opts ...Option) (*Application, error) {
	appConfig := DefaultConfig()
	for _, opt := range opts {
		opt(appConfig)
	}
	return NewFromConfig(*appConfig)
}

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

func (a *Application) Context() context.Context {
	return a.ctx
}

func (a *Application) Config() *config.Config {
	return a.cfg
}

func (a *Application) Logger() *log.Adapter {
	return a.logger
}

func (a *Application) AddRunners(runners ...Runner) {
	a.runners = append(a.runners, runners...)
}

func (a *Application) AddClosers(closers ...Closer) {
	a.closers = append(a.closers, closers...)
}

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

func (a *Application) Shutdown() {
	a.Close()
	a.cancel()
}

func (a *Application) Close() {
	for i := 0; i < len(a.closers); i++ {
		closer := a.closers[i]
		err := closer.Close()
		if err != nil {
			a.logger.Error(a.ctx, errors.WithMessagef(err, "closers[%d] -> %T", i, closer))
		}
	}
}
