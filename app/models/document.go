package models

import (
	"time"
)

type Document struct {
	ID          string            `json:"id" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description,omitempty"`
	File        string            `json:"file,omitempty"` // Path or URL to the file
	Categories  []Category        `json:"categories,omitempty"`
	Content     map[string]string `json:"content,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}

func (Document) TableName() string {
	return "documents"
}
