package mcounter

import "github.com/pkg/errors"

var (
	DuplicateNameErr   = errors.New("duplicate name")
	ContextCanceledErr = errors.New("context canceled")
	InvalidNameErr     = errors.New("name should not contain commas")
	EmptyLabelsErr     = errors.New("empty labels")
)
