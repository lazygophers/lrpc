package mock

import "errors"

var (
	// ErrInvalidArgument is returned when an invalid argument is provided
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrNotImplemented is returned when a method is not implemented in mock
	ErrNotImplemented = errors.New("not implemented")
)
