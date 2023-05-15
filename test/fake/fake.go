package fake

import (
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
)

type Option = options.OptionFunc

func It[T any](opts ...Option) T {
	allOpts := []Option{
		options.WithRandomMapAndSliceMinSize(1),
		options.WithRandomMapAndSliceMaxSize(1),
	}
	allOpts = append(allOpts, opts...)
	var result T
	err := faker.FakeData(&result, allOpts...)
	if err != nil {
		panic(err)
	}
	return result
}
