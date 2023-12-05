package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Adapter struct {
	validator *validator.Validate
}

func New() Adapter {
	return Adapter{
		validator: validator.New(),
	}
}

type wrapper struct {
	V any
}

func (a Adapter) Validate(v any) (ok bool, details map[string]string) {
	err := a.validator.Struct(wrapper{v}) //hack
	if err == nil {
		return true, nil
	}
	details, err = a.collectDetails(err)
	if err != nil {
		return false, map[string]string{"#validator": err.Error()}
	}
	return false, details
}

func (a Adapter) ValidateToError(v any) error {
	ok, details := a.Validate(v)
	if ok {
		return nil
	}
	descriptions := make([]string, 0, len(details))
	for field, err := range details {
		descriptions = append(descriptions, fmt.Sprintf("%s -> %s", field, err))
	}
	err := strings.Join(descriptions, "; ")
	return errors.New(err)
}

func (a Adapter) collectDetails(err error) (map[string]string, error) {
	e, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, err
	}
	const prefixToDelete = "wrapper.V"
	result := make(map[string]string, len(e))
	for _, err := range e {
		field, _ := strings.CutPrefix(err.Namespace(), prefixToDelete)
		result[field] = strings.ReplaceAll(err.Error(), prefixToDelete, "")
	}
	return result, nil
}
