// Package fake provides utilities for generating fake test data using
// the faker library. It offers a simple interface to create populated
// structs for testing.
package fake

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
)

// Option is an alias for faker configuration options.
type Option = options.OptionFunc

// MinSliceSize sets the minimum size for randomly generated slices and maps.
func MinSliceSize(value uint) Option {
	return options.WithRandomMapAndSliceMinSize(value)
}

// MaxSliceSize sets the maximum size for randomly generated slices and maps.
func MaxSliceSize(value uint) Option {
	return options.WithRandomMapAndSliceMaxSize(value)
}

// It generates and returns a fake instance of type T.
// By default, it configures slices and maps to have a size of 1.
// Additional options can be provided to customize the generation behavior.
//
// Panics if the faker library encounters an error during generation.
//
// nolint:ireturn
func It[T any](opts ...Option) T {
	allOpts := []Option{
		MinSliceSize(1),
		MaxSliceSize(1),
	}
	allOpts = append(allOpts, opts...)
	var result T
	err := faker.FakeData(&result, allOpts...)
	if err != nil {
		panic(err)
	}
	return result
}
