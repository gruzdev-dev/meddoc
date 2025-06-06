package models

import (
	"time"
)

type Document struct {
	ID          string            `json:"id" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description,omitempty"`
	Date        string            `json:"date,omitempty"`
	File        string            `json:"file,omitempty"`
	Category    string            `json:"category,omitempty"`
	Priority    int               `json:"priority,omitempty"`
	Content     map[string]string `json:"content,omitempty"`
	UserID      string            `json:"-" binding:"required"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type DocumentCreation struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Date        string            `json:"date,omitempty"`
	File        string            `json:"file,omitempty"`
	Category    string            `json:"category,omitempty"`
	Priority    int               `json:"priority,omitempty"`
	Content     map[string]string `json:"content,omitempty"`
}

type DocumentUpdate struct {
	Title       *string           `json:"title,omitempty"`
	Description *string           `json:"description,omitempty"`
	Date        *string           `json:"date,omitempty"`
	File        *string           `json:"file,omitempty"`
	Category    *string           `json:"category,omitempty"`
	Priority    *int              `json:"priority,omitempty"`
	Content     map[string]string `json:"content,omitempty"`
}
