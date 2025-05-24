package errors

import "errors"

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrAccessDenied     = errors.New("access denied")
)
