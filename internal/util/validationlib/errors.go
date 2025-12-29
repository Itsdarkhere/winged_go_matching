package validationlib

import "errors"

var (
	ErrInvalidNullableString = errors.New("provided nullable string, but empty")
)
