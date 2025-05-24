package domain

import "errors"

var (
	// ErrDatabaseError is returned when a database operation fails
	ErrDatabaseError = errors.New("database error")
)
