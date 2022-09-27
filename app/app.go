package app

import (
	"context"
	"fmt"

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

func New(isDev bool, cfgOpts ...config.Option) (*Application, error) {
	cfg, err := config.New(cfgOpts...)
	if err != nil {
		return nil, errors.WithMessage(err, "create config")
	}

	loggerOpts := []log.Option{log.WithDevelopmentMode(), log.WithLevel(log.DebugLevel)}
	if !isDev {
		loggerOpts = []log.Option{log.WithLevel(log.InfoLevel)}
		logFilePath := cfg.Optional().String("LOGFILE.PATH", "")
		if logFilePath != "" {
			rotation := log.Rotation{
				File:       logFilePath,
				MaxSizeMb:  cfg.Optional().Int("LOGFILE.MAXSIZEMB", 512),
				MaxDays:    0,
				MaxBackups: cfg.Optional().Int("LOGFILE.MAXBACKUPS", 4),
				Compress:   cfg.Optional().Bool("LOGFILE.COMPRESS", true),
			}
			loggerOpts = append(loggerOpts, log.WithFileRotation(rotation))
		}
	}
	logger, err := log.New(loggerOpts...)
	if err != nil {
		return nil, errors.WithMessage(err, "create logger")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Application{
		ctx:     ctx,
		cfg:     cfg,
		logger:  logger,
		closers: []Closer{logger},
		cancel:  cancel,
	}, nil
}

func (a Application) Context() context.Context {
	return a.ctx
}

func (a Application) Config() *config.Config {
	return a.cfg
}

func (a Application) Logger() *log.Adapter {
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
		go func(runner Runner) {
			err := runner.Run(a.ctx)
			if err != nil {
				select {
				case errChan <- errors.WithMessagef(err, "start runner[%T]", runner):
				default:
					a.logger.Error(a.ctx, errors.WithMessagef(err, "start runner[%T]", runner))
				}
			}
		}(a.runners[i])
	}

	select {
	case err := <-errChan:
		return err
	case <-a.ctx.Done():
		return nil
	}
}

func (a *Application) Shutdown() {
	a.cancel()

	for i := 0; i < len(a.closers); i++ {
		closer := a.closers[i]
		err := closer.Close()
		if err != nil {
			a.logger.Error(a.ctx, err, log.String("closer", fmt.Sprintf("%T", closer)))
		}
	}
}
