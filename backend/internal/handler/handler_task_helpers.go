package handler

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
)

type jsonUnmarshalTypeError = json.UnmarshalTypeError

func asJSONTypeError(err error, target **json.UnmarshalTypeError) bool {
	return errors.As(err, target)
}

func asInvalidValidation(err error, target **validator.InvalidValidationError) bool {
	return errors.As(err, target)
}

func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	return errors.As(err, target)
}
