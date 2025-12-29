package validationlib

import (
	"errors"
	"fmt"
	"strings"
	"wingedapp/pgtester/internal/util/errutil"
	"wingedapp/pgtester/internal/util/interfaceutil"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

/*
	TODO: rewrite this library to be more idiomatic, and less sloppy
*/

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
	trans    ut.Translator
)

var (
	ErrStructNil = errors.New("error, struct is nil")
)

func init() {
	configValidate()
}

func ValidateAsErr(p interface{}, exclusions ...string) error {
	v := validate_(p)
	return v.Error()
}

func ValidateAsArr(p interface{}, exclusions ...string) errutil.List {
	return validate_(p)
}

func validate_(p interface{}, exclusions ...string) errutil.List {
	if p != nil && fmt.Sprintf("%v", p) == "<nil>" {
		return errutil.NewFromError(ErrStructNil)
	}

	if p == nil {
		return errutil.NewFromError(ErrStructNil)
	}

	var errList errutil.List

	structVal, err := interfaceutil.GetStructAsValue(p)
	if err != nil {
		errList.AddErr(err)
		return errList
	}

	err = validate.Struct(structVal)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, err := range validationErrors {
			translatedErr := err.Translate(trans)
			if strings.Contains(translatedErr, "__VAL__") {
				val, ok := err.Value().(string)
				if ok {
					translatedErr = strings.ReplaceAll(translatedErr, "__VAL__", val)
				}
			}
			if translatedErr != "" {
				errList = append(errList, errors.New(translatedErr))
			} else {
				errList = append(errList, err)
			}
		}
	}

	return errList
}

func Validate(p interface{}, exclusions ...string) error {
	if p != nil && fmt.Sprintf("%v", p) == "<nil>" {
		return ErrStructNil
	}

	if p == nil {
		return ErrStructNil
	}

	var errList errutil.List

	structVal, err := interfaceutil.GetStructAsValue(p)
	if err != nil {
		errList.AddErr(err)
		return errList.Single()
	}

	err = validate.Struct(structVal)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, err := range validationErrors {
			translatedErr := err.Translate(trans)
			if strings.Contains(translatedErr, "__VAL__") {
				val, ok := err.Value().(string)
				if ok {
					translatedErr = strings.ReplaceAll(translatedErr, "__VAL__", val)
				}
			}
			if translatedErr != "" {
				errList = append(errList, errors.New(translatedErr))
			} else {
				errList = append(errList, err)
			}
		}
	}

	return errList.Single()
}
