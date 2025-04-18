package domain

import "errors"

var (
	// ErrDocumentNotFound is returned when a document is not found
	ErrDocumentNotFound = errors.New("document not found")

	// ErrInvalidDocumentType is returned when document type is invalid
	ErrInvalidDocumentType = errors.New("invalid document type")

	// ErrInvalidPatientID is returned when patient ID is invalid
	ErrInvalidPatientID = errors.New("invalid patient ID")

	// ErrEmptyContent is returned when document content is empty
	ErrEmptyContent = errors.New("document content cannot be empty")

	// ErrDatabaseError is returned when a database operation fails
	ErrDatabaseError = errors.New("database error")

	// ErrEmptyTitle is returned when the title of a document is empty
	ErrEmptyTitle = errors.New("empty title")
)
