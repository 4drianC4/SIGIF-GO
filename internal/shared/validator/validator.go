package validator

import (
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/es"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	esTranslations "github.com/go-playground/validator/v10/translations/es"
	"errors"
	"reflect"
	"strings"
)

type Validator struct {
	validate   *validator.Validate
	translator ut.Translator
}

func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	enLocale := en.New()
	esLocale := es.New()
	uni := ut.New(enLocale, esLocale)
	trans, _ := uni.GetTranslator("es")

	_ = enTranslations.RegisterDefaultTranslations(v, trans)
	_ = esTranslations.RegisterDefaultTranslations(v, trans)

	return &Validator{
		validate:   v,
		translator: trans,
	}
}

func (v *Validator) Validate(s any) error {
	if err := v.validate.Struct(s); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			details := make(map[string]string)
			for _, e := range validationErrors {
				details[e.Field()] = e.Translate(v.translator)
			}
			appErr := sharedErrors.Wrap(err, sharedErrors.CodeValidation, "validation failed")
			appErr.WithDetails(details)
			return appErr
		}
		return sharedErrors.Wrap(err, sharedErrors.CodeValidation, "validation failed")
	}
	return nil
}

func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}

func (v *Validator) RegisterValidation(tag string, fn validator.Func, callValidationEvenIfNull bool) error {
	return v.validate.RegisterValidation(tag, fn, callValidationEvenIfNull)
}

func (v *Validator) RegisterStructValidation(fn validator.StructLevelFunc, types ...any) {
	v.validate.RegisterStructValidation(fn, types...)
}