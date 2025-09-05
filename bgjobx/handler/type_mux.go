package handler

import (
	"context"

	"github.com/txix-open/bgjob"
)

type Mux struct {
	data map[string]SyncHandlerAdapter
}

func NewMux() *Mux {
	return &Mux{
		data: make(map[string]SyncHandlerAdapter),
	}
}

func (m *Mux) Register(jobType string, handler SyncHandlerAdapter) *Mux {
	m.data[jobType] = handler
	return m
}

func (m *Mux) Handle(ctx context.Context, job bgjob.Job) Result {
	handler, ok := m.data[job.Type]
	if !ok {
		return MoveToDlq(bgjob.ErrUnknownType)
	}
	return handler.Handle(ctx, job)
}
