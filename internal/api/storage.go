package api

import (
	"context"

	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

type Storage interface {
	CreateCartItem(ctx context.Context, item models.CartItem) (models.CartItem, error)
	GetCart(ctx context.Context, userID int64) ([]models.CartItem, error)
	DeleteCartItem(ctx context.Context, id int64) error
	CreateField(ctx context.Context, field models.Field) (models.Field, error)
	GetField(ctx context.Context, id int64) (models.Field, error)
	GetFields(ctx context.Context) ([]models.Field, error)
	DeleteField(ctx context.Context, id int64) error
	UpdateField(ctx context.Context, field models.Field) (models.Field, error)
	CreateOption(ctx context.Context, option models.FieldOption) (models.FieldOption, error)
	DeleteOption(ctx context.Context, id int64) error
	UpdateOption(ctx context.Context, option models.FieldOption) (models.FieldOption, error)
	CreateProduct(ctx context.Context, product models.Product) (models.Product, error)
	GetProduct(ctx context.Context, id int64) (models.Product, error)
	GetProducts(ctx context.Context, limit int, offset int) ([]models.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
	UpdateProduct(ctx context.Context, product models.Product) (models.Product, error)
	CreateTemplate(ctx context.Context, template models.Template) (models.Template, error)
	GetTemplate(ctx context.Context, id int64) (models.Template, error)
	GetTemplates(ctx context.Context, limit int, offset int) ([]models.Template, error)
	DeleteTemplate(ctx context.Context, id int64) error
	UpdateTemplate(ctx context.Context, template models.Template) (models.Template, error)
	GetUserByTelegramID(ctx context.Context, userTelegramID int64) (*models.User, error)
}

var _ Storage = (*storage.Storage)(nil)
