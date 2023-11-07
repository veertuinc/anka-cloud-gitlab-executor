package gitlab

import (
	"errors"
	"fmt"
)

var ErrMissingVar = errors.New("missing environment variable")

// only use for errors.Is() calls. For wrapping, use TransientError().
var ErrTransient = fmt.Errorf("transient error")

type transientError struct {
	e error
}

func (e transientError) Error() string {
	return e.e.Error()
}

func (e transientError) Unwrap() error {
	return e.e
}

func (e transientError) Is(target error) bool {
	return target == ErrTransient
}

func TransientError(e error) error {
	return transientError{e: e}
}
