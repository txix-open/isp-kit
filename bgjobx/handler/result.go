package handler

import (
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

// Result defines the outcome of job processing.
// One of the boolean flags (Complete, Retry, MoveToDlq, Reschedule, RescheduleWithArg)
// should be set to indicate the desired action.
type Result struct {
	// Complete indicates the job was successfully processed.
	Complete bool
	// Err contains the error that occurred during processing.
	Err error
	// MoveToDlq indicates the job should be moved to the dead letter queue.
	MoveToDlq bool
	// Retry indicates the job should be retried.
	Retry bool
	// RetryDelay specifies the delay before retrying the job.
	RetryDelay time.Duration

	// Reschedule indicates the job should be rescheduled.
	Reschedule bool
	// RescheduleDelay specifies when to reschedule the job.
	RescheduleDelay time.Duration
	// RescheduleWithArg indicates the job should be rescheduled with a new argument.
	RescheduleWithArg bool
	// Arg contains the new payload for rescheduled jobs.
	Arg []byte
}

// RescheduleBy is an interface for types that can calculate a rescheduling delay.
type RescheduleBy interface {
	Reschedule() (time.Duration, error)
}

// Complete returns a Result indicating successful job completion.
func Complete() Result {
	return Result{Complete: true}
}

// Retry returns a Result indicating the job should be retried after the specified delay.
func Retry(after time.Duration, err error) Result {
	return Result{Retry: true, RetryDelay: after, Err: err}
}

// MoveToDlq returns a Result indicating the job should be moved to the dead letter queue.
func MoveToDlq(err error) Result {
	return Result{MoveToDlq: true, Err: err}
}

// Cron implements RescheduleBy using a cron expression to calculate the next run time.
type Cron struct {
	cronExpression string
	currentTime    time.Time
}

// ByCron creates a Cron reschedule option using a cron expression and current time.
func ByCron(cronExpression string, currentTime time.Time) Cron {
	return Cron{
		cronExpression: cronExpression,
		currentTime:    currentTime,
	}
}

// Reschedule calculates the duration until the next scheduled time based on the cron expression.
func (b Cron) Reschedule() (time.Duration, error) {
	return rescheduleByCron(b.cronExpression, b.currentTime)
}

// rescheduleByCron parses the cron expression and returns the duration until the next run.
func rescheduleByCron(cronExpression string, currentTime time.Time) (time.Duration, error) {
	schedule, err := cron.ParseStandard(cronExpression)
	if err != nil {
		return 0, err
	}
	nextRunAt := schedule.Next(currentTime)
	return nextRunAt.Sub(currentTime), nil
}

// AfterTime implements RescheduleBy for rescheduling after a fixed duration.
type AfterTime struct {
	after       time.Duration
	currentTime time.Time
}

// ByAfterTime creates an AfterTime reschedule option.
func ByAfterTime(after time.Duration, currentTime time.Time) AfterTime {
	return AfterTime{
		after:       after,
		currentTime: currentTime,
	}
}

// Reschedule returns the fixed duration for rescheduling.
func (b AfterTime) Reschedule() (time.Duration, error) {
	return b.after, nil
}

// RescheduleOption is a function type for configuring reschedule options.
type RescheduleOption func(opt *rescheduleOptions)

type rescheduleOptions struct {
	Arg               []byte
	RescheduleWithArg bool
}

// WithArg configures a reschedule to include a new payload.
func WithArg(arg []byte) RescheduleOption {
	return func(opt *rescheduleOptions) {
		if len(arg) > 0 {
			opt.Arg = arg
			opt.RescheduleWithArg = true
		}
	}
}

// Reschedule creates a Result for rescheduling a job at a future time.
// It uses the provided RescheduleBy implementation to calculate the delay.
// Optional RescheduleOption functions can be used to include a new payload.
// If the reschedule calculation fails, the job is moved to the DLQ.
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
