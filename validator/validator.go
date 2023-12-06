package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
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

func (a Adapter) collectDetails(err error) (map[string]string, error) {
	e, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, err
	}
	const prefixToDelete = "wrapper.V"
	result := make(map[string]string, len(e))
	for _, err := range e {
		field, _ := strings.CutPrefix(err.Namespace(), prefixToDelete)
		result[field] = err.Translate(a.translator)
	}
	return result, nil
}

func (a Adapter) ValidateOld(v any) (ok bool, details map[string]string) {
	ok, err := govalidator.ValidateStruct(wrapper{v}) //hack
	if ok || err == nil {
		return true, nil
	}

	details = make(map[string]string)
	err = a.collectDetailsOld(err, details)
	if err != nil {
		return false, map[string]string{"#validator": err.Error()}
	}
	return false, details

}

func (a Adapter) collectDetailsOld(err error, result map[string]string) error {
	switch e := err.(type) {
	case govalidator.Error:
		errName := e.Name
		if len(e.Path) > 0 {
			errName = strings.Join(append(e.Path, e.Name), ".")
			errName = errName[2:] //remove V.
		}
		result[errName] = e.Err.Error()
	case govalidator.Errors:
		for _, err := range e.Errors() {
			err = a.collectDetailsOld(err, result)
			if err != nil {
				return err
			}
		}
	default:
		return err
	}
	return nil
}
