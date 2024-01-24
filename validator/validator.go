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
		for i := 0; i < len(field); i++ {
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
