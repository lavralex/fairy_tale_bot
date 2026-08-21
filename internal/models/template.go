package models

import (
	"time"
)

type Template struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	Fields    []TemplateField
}

type TemplateField struct {
	Field     Field
	SortOrder int
}

type Field struct {
	ID        int64
	Name      string
	IsActive  bool
	CreatedAt time.Time
	Options   []FieldOption
}

type FieldOption struct {
	ID        int64
	Name      string
	FieldID   int64
	IsActive  bool
	SortOrder int
	CreatedAt time.Time
}
