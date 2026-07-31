// Package apperror contains transport-independent application errors.
package apperror

import "fmt"

type Error struct {
	Status  int
	Code    string
	Message string
	Details any
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Cause)
	}
	return e.Code
}
func (e *Error) Unwrap() error { return e.Cause }

func New(status int, code, message string, details any) *Error {
	return &Error{Status: status, Code: code, Message: message, Details: details}
}
func Wrap(cause error) *Error {
	return &Error{Status: 500, Code: "INTERNAL_ERROR", Message: "An unexpected error occurred.", Cause: cause}
}
func Validation(field, message string) *Error {
	return New(400, "VALIDATION_ERROR", "The request contains invalid fields.", []map[string]string{{"field": field, "message": message}})
}
