package handler

import (
	"context"

	"github.com/txix-open/bgjob"
)

// Mux routes jobs to different handlers based on their type.
// It maintains a registry of job types and their corresponding handlers,
// returning an error for unknown job types.
type Mux struct {
	data map[string]SyncHandlerAdapter
}

// NewMux creates a new empty Mux instance.
func NewMux() *Mux {
	return &Mux{
		data: make(map[string]SyncHandlerAdapter),
	}
}

// Register associates a job type with its handler.
// Returns the Mux for method chaining.
func (m *Mux) Register(jobType string, handler SyncHandlerAdapter) *Mux {
	m.data[jobType] = handler
	return m
}

// Handle routes a job to the appropriate handler based on its type.
// If no handler is registered for the job type, it returns MoveToDlq
// with ErrUnknownType.
func (m *Mux) Handle(ctx context.Context, job bgjob.Job) Result {
	handler, ok := m.data[job.Type]
	if !ok {
		return MoveToDlq(bgjob.ErrUnknownType)
	}
	return handler.Handle(ctx, job)
}
