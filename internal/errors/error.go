package errors

import "fmt"

// add context to error
func Wrap(baseErr error, msg string) error {
	return fmt.Errorf("%s: %w", msg, baseErr)
}
