package errors

import "fmt"

type ErrMissingEnvVar struct {
	name string
}

func (e *ErrMissingEnvVar) Error() string {
	return fmt.Sprintf("missing environment variable: %s", e.name)
}

func MissingEnvVar(name string) error {
	return &ErrMissingEnvVar{name: name}
}
