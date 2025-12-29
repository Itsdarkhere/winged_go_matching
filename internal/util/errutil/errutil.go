package errutil

import (
	"errors"
	"sync"
)

type Util struct {
	StatusCode int
	List       List
}

func (e *Util) Error() string {
	return e.List.Single().Error()
}

type Cfg struct {
	StatusCode int
	Err        error
	Errs       List
}

func NewFromError(err error) List {
	var errList List
	errList.AddErr(err)
	return errList
}

type List []error

var m sync.Mutex

func (e *List) AddErr(err error) {
	m.Lock()
	defer m.Unlock()
	ensureErrList(e)
	*e = append(*e, err)
}

// ErrAsUtil checks if the error is an underlying
// *Util object.
func ErrAsUtil(i interface{}) (*Util, bool) {
	_, ok := i.(error)
	if !ok {
		return nil, false
	}

	var util *Util
	ok = errors.As(i.(error), &util)
	if !ok {
		return nil, false
	}

	return util, true
}

// AsErrUtil returns the List alongside a statusCode,
// very useful if we want to add more context on what FakeAPI
func (e *List) AsErrUtil(statusCode int) error {
	return &Util{
		List:       *e,
		StatusCode: statusCode,
	}
}

func ensureErrList(e *List) {
	if e == nil {
		*e = make([]error, 0)
	}
}

func (e *List) Add(err string) {
	m.Lock()
	defer m.Unlock()
	ensureErrList(e)
	*e = append(*e, errors.New(err))
}

func (e *List) Wrapper() Wrapper {
	return Wrapper{list: *e}
}

func (e *List) Error() error {
	if e == nil {
		return nil
	}
	if len(*e) == 0 {
		return nil
	}

	return newWrapper(*e)
}

func (e *List) HasError(i string) bool {
	var found bool
	return found
}

// Single will return the error string value
func (e *List) Single() error {
	m.Lock()
	defer m.Unlock()

	ensureErrList(e)

	deref := *e

	switch len(deref) {
	case 0:
		return nil

	case 1:
		return deref[0]
	}

	var bs []byte
	for _, err := range deref {
		if err == nil {
			continue
		}
		bs = append(bs, []byte(err.Error())...)
		bs = append(bs, ',', ' ')
	}

	return errors.New(string(bs))
}

func (e *List) HasErrors() bool {
	if e == nil {
		return false
	}
	return len(*e) != 0
}

func (e *List) ErrsAsStrArr() []string {
	if e == nil {
		return nil
	}
	arr := make([]string, 0)
	for _, v := range *e {
		arr = append(arr, v.Error())
	}
	return arr
}

// ValidationError is a sentinel error type that business layers can return
// to signal validation failures with detailed error messages.
// The API layer will detect this type and convert it to a proper response
// with the errors array populated.
type ValidationError struct {
	Message string
	Errors  []string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return e.Message
	}
	return e.Message + ": " + e.Errors[0]
}

// NewValidationError creates a new ValidationError with a message and error list.
func NewValidationError(message string, errs []string) *ValidationError {
	if message == "" {
		message = "validation failed"
	}
	return &ValidationError{
		Message: message,
		Errors:  errs,
	}
}

// NewValidationErrorFromList creates a ValidationError from a List.
func NewValidationErrorFromList(message string, errs *List) *ValidationError {
	if message == "" {
		message = "validation failed"
	}
	return &ValidationError{
		Message: message,
		Errors:  errs.ErrsAsStrArr(),
	}
}
