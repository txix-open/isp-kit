// Package validator provides a structured validation adapter using the
// go-playground/validator library with English translations for error messages.
//
// The package wraps the underlying validator to provide a simple API that returns
// validation results as a boolean flag and a map of field-specific error messages.
// It also provides a convenience method to convert validation failures to a single
// error with formatted field descriptions.
//
// Example usage:
//
//	adapter := validator.New()
//	ok, details := adapter.Validate(myStruct)
//	if !ok {
//	    // details contains field -> error message mappings
//	}
package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/txix-open/validator/v10"
	en_translations "github.com/txix-open/validator/v10/translations/en"
)

// Adapter provides validation functionality with translated error messages.
// It wraps the go-playground/validator library and handles translation of
// validation errors into user-friendly messages.
type Adapter struct {
	validator  *validator.Validate
	translator ut.Translator
}

// New creates a new Adapter instance with English translation support.
// The returned Adapter is safe for concurrent use.
func New() Adapter {
	enTranslator := en.New()
	uni := ut.New(enTranslator, enTranslator)
	translator, _ := uni.GetTranslator("en")
	validator := validator.New()
	err := en_translations.RegisterDefaultTranslations(validator, translator)
	if err != nil {
		panic(err)
	}
	return Adapter{
		validator:  validator,
		translator: translator,
	}
}

// wrapper is an internal type used to avoid validating the Adapter struct itself.
type wrapper struct {
	V any
}

// Validate checks if the provided value passes all validation rules.
// It returns true and nil details if validation succeeds, or false and a map
// of field names to error messages if validation fails.
// The value should be a struct with validator tags defined on its fields.
// Returns an error in the details map under "#validator" key if an unexpected
// error occurs during validation.
func (a Adapter) Validate(v any) (ok bool, details map[string]string) {
	err := a.validator.Struct(wrapper{v}) // hack
	if err == nil {
		return true, nil
	}
	details, err = a.collectDetails(err)
	if err != nil {
		return false, map[string]string{"#validator": err.Error()}
	}
	return false, details
}

// ValidateToError validates the provided value and returns an error if validation fails.
// Returns nil if validation succeeds.
// The error message contains semicolon-separated field -> error message pairs
// for all validation failures.
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
	return errors.New(err) //nolint:err113
}

const (
	prefixToDelete = "wrapper.V"
)

// collectDetails extracts field-specific error messages from a validation error.
// It transforms the validation error namespace to field names and translates
// the error messages using the configured translator.
func (a Adapter) collectDetails(err error) (map[string]string, error) {
	var e validator.ValidationErrors
	if !errors.As(err, &e) {
		return nil, err
	}
	result := make(map[string]string, len(e))
	for _, err := range e {
		field := []byte(err.Namespace())[len(prefixToDelete):]
		if field[0] == '.' {
			field = field[1:]
		}
		firstLetter := 0
		for i := range field {
			if field[i] == '.' {
				field[firstLetter] = strings.ToLower(string(field[firstLetter]))[0]
				firstLetter = i + 1
			}
		}
		field[firstLetter] = strings.ToLower(string(field[firstLetter]))[0]
		result[string(field)] = err.Translate(a.translator)
	}
	return result, nil
}
