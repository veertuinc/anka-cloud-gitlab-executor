package gitlab

import (
	"errors"
	"fmt"
)

var ErrMissingVar = errors.New("missing environment variable")
var ErrInvalidVar = errors.New("invalid environment variable")

var ErrTransient = errors.New("") // message is empty to avoid showing unnecessary information to the user

func TransientError(e error) error {
	return fmt.Errorf("%w%w", ErrTransient, e)
}
