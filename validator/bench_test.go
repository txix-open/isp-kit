package validator_test

import (
	"strings"
	"testing"

	"github.com/asaskevich/govalidator"
	"gitlab.txix.ru/isp/isp-kit/validator"
)

type inner struct {
	Str string `validate:"required,min=4" valid:"required,minstringlength(4)"`
}

type bench struct {
	Slice []inner `validate:"required" valid:"required"`
	Int   int     `validate:"required,min=1,max=5" valid:"required,range(1|5)"`
	Str   string  `validate:"required,oneof=debug info error" valid:"required,in(debug|info|error)"`
}

func BenchmarkValidatorV10(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		obj := getBenchData()
		for pb.Next() {
			validator.Default.Validate(obj)
		}
	})
}

func BenchmarkValidatorAsaskevich(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		obj := getBenchData()
		for pb.Next() {
			validateAsaskevich(obj)
		}
	})
}

func getBenchData() *bench {
	return &bench{
		Slice: []inner{
			{Str: "config"},
			{Str: "client"},
			{Str: "endpoints"},
			{Str: "app"},
		},
		Int: 5,
		Str: "info",
	}
}

type wrapper struct {
	V any
}

func validateAsaskevich(v any) (ok bool, details map[string]string) {
	ok, err := govalidator.ValidateStruct(wrapper{v}) // hack
	if ok || err == nil {
		return true, nil
	}
	details = make(map[string]string)
	err = collectDetails(err, details)
	if err != nil {
		return false, map[string]string{"#validator": err.Error()}
	}
	return false, details

}

func collectDetails(err error, result map[string]string) error {
	switch e := err.(type) {
	case govalidator.Error:
		errName := e.Name
		if len(e.Path) > 0 {
			errName = strings.Join(append(e.Path, e.Name), ".")
			errName = errName[2:] // remove V.
		}
		result[errName] = e.Err.Error()
	case govalidator.Errors:
		for _, err := range e.Errors() {
			err = collectDetails(err, result)
			if err != nil {
				return err
			}
		}
	default:
		return err
	}
	return nil
}
