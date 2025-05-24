package models

import (
	"time"
)

type Category struct {
	ID          string    `json:"id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}
