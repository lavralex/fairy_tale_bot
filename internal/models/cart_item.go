package models

import (
	"time"
)

type CartItem struct {
	ID        int64
	UserID    int64
	ProductID int64
	Price     string
	Name      string
	Comment   *string
	CreatedAt time.Time
	Options   []CartItemOption
}

type CartItemOption struct {
	ID         int64
	CartItemID int64
	FieldID    int64
	FieldName  string
	OptionID   int64
	Name       string
	CreatedAt  time.Time
}
