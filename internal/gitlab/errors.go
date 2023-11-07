package gitlab

import (
	"errors"
	"fmt"
)

var ErrTransient = fmt.Errorf("") // this error is used to signal that the error is transient and can be retried

func TransientError(err error) error {
	return fmt.Errorf("%w%w", ErrTransient, err)
}

var ErrMissingVar = errors.New("missing environment variable")
