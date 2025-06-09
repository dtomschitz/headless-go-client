package internal

import "errors"

type (
	NotFoundError struct {
		err error
	}

	InvalidRequestError struct {
		err error
	}

	ConflictError struct {
		err error
	}
)

func NewNotFoundError(err error) *NotFoundError {
	return &NotFoundError{err: err}
}

func (e *NotFoundError) Error() string {
	return e.err.Error()
}

func IsNotFoundError(err error) bool {
	var notFoundError *NotFoundError
	return errors.As(err, &notFoundError)
}

func NewInvalidRequestError(err error) *InvalidRequestError {
	return &InvalidRequestError{err: err}
}

func (e *InvalidRequestError) Error() string {
	return e.err.Error()
}

func IsInvalidRequestError(err error) bool {
	var invalidRequestError *InvalidRequestError
	return errors.As(err, &invalidRequestError)
}

func NewConflictError(err error) *ConflictError {
	return &ConflictError{err: err}
}

func (e *ConflictError) Error() string {
	return e.err.Error()
}

func IsConflictError(err error) bool {
	var conflictError *ConflictError
	return errors.As(err, &conflictError)
}
