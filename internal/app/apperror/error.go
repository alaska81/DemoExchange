package apperror

import "fmt"

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e
}

func (e *Error) Wrap(err error) error {
	return fmt.Errorf("%v: %w", e, err)
}

func New(message string) *Error {
	return &Error{
		Message: message,
	}
}
