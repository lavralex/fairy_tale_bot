package models

import (
	"time"
)

type Product struct {
	ID              int64
	Name            string
	Description     string
	Price           string
	TemplateID      *int64
	Tags            []string
	CreatedAt       time.Time
	IsAvailable     bool
	CommentEnabled  bool
	TemplateEnabled bool
	Images          []ProductImage
}

type ProductImage struct {
	ID        int64
	ProductID int64
	Src       string
	Alt       string
	SortOrder int8
	CreatedAt time.Time
}
