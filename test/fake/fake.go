package fake

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
)

type Option = options.OptionFunc

func MinSliceSize(value uint) Option {
	return options.WithRandomMapAndSliceMinSize(value)
}

func MaxSliceSize(value uint) Option {
	return options.WithRandomMapAndSliceMaxSize(value)
}

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
