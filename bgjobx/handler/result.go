package handler

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"time"
)

type Result struct {
	Complete   bool
	Err        error
	MoveToDlq  bool
	Retry      bool
	RetryDelay time.Duration

	Reschedule        bool
	RescheduleDelay   time.Duration
	RescheduleWithArg bool
	Arg               []byte
}

type RescheduleBy interface {
	Reschedule() (time.Duration, error)
}

func Complete() Result {
	return Result{Complete: true}
}

func Retry(after time.Duration, err error) Result {
	return Result{Retry: true, RetryDelay: after, Err: err}
}

func MoveToDlq(err error) Result {
	return Result{MoveToDlq: true, Err: err}
}

type Cron struct {
	cronExpression string
	currentTime    time.Time
}

func ByCron(cronExpression string, currentTime time.Time) Cron {
	return Cron{
		cronExpression: cronExpression,
		currentTime:    currentTime,
	}
}

func (b Cron) Reschedule() (time.Duration, error) {
	return rescheduleByCron(b.cronExpression, b.currentTime)
}

func rescheduleByCron(cronExpression string, currentTime time.Time) (time.Duration, error) {
	schedule, err := cron.ParseStandard(cronExpression)
	if err != nil {
		return 0, err
	}
	nextRunAt := schedule.Next(currentTime)
	return nextRunAt.Sub(currentTime), nil
}

type AfterTime struct {
	after       time.Duration
	currentTime time.Time
}

func ByAfterTime(after time.Duration, currentTime time.Time) AfterTime {
	return AfterTime{
		after:       after,
		currentTime: currentTime,
	}
}

func (b AfterTime) Reschedule() (time.Duration, error) {
	return b.after, nil
}

type RescheduleOption func(opt *rescheduleOptions)

type rescheduleOptions struct {
	Arg               []byte
	RescheduleWithArg bool
}

func WithArg(arg []byte) RescheduleOption {
	return func(opt *rescheduleOptions) {
		if len(arg) > 0 {
			opt.Arg = arg
			opt.RescheduleWithArg = true
		}
	}
}

func Reschedule(by RescheduleBy, opts ...RescheduleOption) Result {
	options := &rescheduleOptions{}
	for _, opt := range opts {
		opt(options)
	}

	rescheduleDelay, err := by.Reschedule()
	if err != nil {
		return MoveToDlq(errors.WithMessage(err, "failed to reschedule"))
	}

	return Result{
		Reschedule:        !options.RescheduleWithArg,
		RescheduleWithArg: options.RescheduleWithArg,
		Arg:               options.Arg,
		RescheduleDelay:   rescheduleDelay,
	}
}
