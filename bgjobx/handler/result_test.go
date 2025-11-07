package handler_test

import (
	"github.com/txix-open/isp-kit/bgjobx/handler"
	"github.com/txix-open/isp-kit/test"
	"testing"
	"time"
)

func TestReschedule_ByAfterTime(t *testing.T) {
	t.Parallel()

	_, assert := test.New(t)

	baseTime := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := handler.Result{
		Reschedule: true,
		NextRunAt:  baseTime.Add(24 * time.Hour),
	}

	result := handler.Reschedule(handler.ByAfterTime(24*time.Hour, baseTime))

	assert.Equal(expected, result)
}

func TestReschedule_ByCron(t *testing.T) {
	t.Parallel()

	_, assert := test.New(t)

	testTime := time.Date(2001, 01, 01, 0, 0, 0, 0, time.UTC)
	cronExpression := "0 0 * * *"
	expected := handler.Result{
		Reschedule: true,
		NextRunAt:  time.Date(2001, 01, 02, 0, 0, 0, 0, time.UTC),
	}

	result := handler.Reschedule(handler.ByCron(cronExpression, testTime))

	assert.Equal(expected, result)
}

func TestReschedule_WithArg(t *testing.T) {
	t.Parallel()

	_, assert := test.New(t)

	testTime := time.Date(2001, 01, 01, 0, 0, 0, 0, time.UTC)
	expected := handler.Result{
		RescheduleWithArg: true,
		Arg:               []byte("test"),
		NextRunAt:         time.Date(2001, 01, 02, 0, 0, 0, 0, time.UTC),
	}

	result := handler.Reschedule(handler.ByAfterTime(24*time.Hour, testTime), handler.WithArg([]byte("test")))

	assert.Equal(expected, result)
}
