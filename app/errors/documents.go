package errors

import "errors"

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrInternal         = errors.New("internal server error")
)
