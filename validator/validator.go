package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/integration-system/validator/v10"
	en_translations "github.com/integration-system/validator/v10/translations/en"
)

type Adapter struct {
	validator  *validator.Validate
	translator ut.Translator
}

func New() Adapter {
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")
	validator := validator.New()
	_ = en_translations.RegisterDefaultTranslations(validator, translator)
	return Adapter{
		validator:  validator,
		translator: translator,
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

const (
	prefixToDelete = "wrapper.V"
)

func (a Adapter) collectDetails(err error) (map[string]string, error) {
	e, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, err
	}
	result := make(map[string]string, len(e))
	for _, err := range e {
		field := err.Namespace()[len(prefixToDelete):]
		result[field] = err.Translate(a.translator)
	}
	return result, nil
}
