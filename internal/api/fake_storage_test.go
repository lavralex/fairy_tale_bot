package api

import (
	"context"

	"github.com/lavralex/fairy_tale_bot/internal/models"
)

type fakeStorage struct {
	createCartItemFunc      func(ctx context.Context, item models.CartItem) (models.CartItem, error)
	getCartFunc             func(ctx context.Context, userID int64) ([]models.CartItem, error)
	deleteCartItemFunc      func(ctx context.Context, id int64) error
	createFieldFunc         func(ctx context.Context, field models.Field) (models.Field, error)
	getFieldFunc            func(ctx context.Context, id int64) (models.Field, error)
	getFieldsFunc           func(ctx context.Context) ([]models.Field, error)
	deleteFieldFunc         func(ctx context.Context, id int64) error
	updateFieldFunc         func(ctx context.Context, field models.Field) (models.Field, error)
	createOptionFunc        func(ctx context.Context, option models.FieldOption) (models.FieldOption, error)
	deleteOptionFunc        func(ctx context.Context, id int64) error
	updateOptionFunc        func(ctx context.Context, option models.FieldOption) (models.FieldOption, error)
	createProductFunc       func(ctx context.Context, product models.Product) (models.Product, error)
	getProductFunc          func(ctx context.Context, id int64) (models.Product, error)
	getProductsFunc         func(ctx context.Context, limit int, offset int) ([]models.Product, error)
	deleteProductFunc       func(ctx context.Context, id int64) error
	updateProductFunc       func(ctx context.Context, product models.Product) (models.Product, error)
	createTemplateFunc      func(ctx context.Context, template models.Template) (models.Template, error)
	getTemplateFunc         func(ctx context.Context, id int64) (models.Template, error)
	getTemplatesFunc        func(ctx context.Context, limit int, offset int) ([]models.Template, error)
	deleteTemplateFunc      func(ctx context.Context, id int64) error
	updateTemplateFunc      func(ctx context.Context, template models.Template) (models.Template, error)
	getUserByTelegramIDFunc func(ctx context.Context, userTelegramID int64) (*models.User, error)
}

func (f *fakeStorage) CreateCartItem(ctx context.Context, item models.CartItem) (models.CartItem, error) {
	return f.createCartItemFunc(ctx, item)
}

func (f *fakeStorage) GetCart(ctx context.Context, userID int64) ([]models.CartItem, error) {
	return f.getCartFunc(ctx, userID)
}

func (f *fakeStorage) DeleteCartItem(ctx context.Context, id int64) error {
	return f.deleteCartItemFunc(ctx, id)
}

func (f *fakeStorage) CreateField(ctx context.Context, field models.Field) (models.Field, error) {
	return f.createFieldFunc(ctx, field)
}

func (f *fakeStorage) GetField(ctx context.Context, id int64) (models.Field, error) {
	return f.getFieldFunc(ctx, id)
}

func (f *fakeStorage) GetFields(ctx context.Context) ([]models.Field, error) {
	return f.getFieldsFunc(ctx)
}

func (f *fakeStorage) DeleteField(ctx context.Context, id int64) error {
	return f.deleteFieldFunc(ctx, id)
}

func (f *fakeStorage) UpdateField(ctx context.Context, field models.Field) (models.Field, error) {
	return f.updateFieldFunc(ctx, field)
}

func (f *fakeStorage) CreateOption(ctx context.Context, option models.FieldOption) (models.FieldOption, error) {
	return f.createOptionFunc(ctx, option)
}

func (f *fakeStorage) DeleteOption(ctx context.Context, id int64) error {
	return f.deleteOptionFunc(ctx, id)
}

func (f *fakeStorage) UpdateOption(ctx context.Context, option models.FieldOption) (models.FieldOption, error) {
	return f.updateOptionFunc(ctx, option)
}

func (f *fakeStorage) CreateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	return f.createProductFunc(ctx, product)
}

func (f *fakeStorage) GetProduct(ctx context.Context, id int64) (models.Product, error) {
	return f.getProductFunc(ctx, id)
}

func (f *fakeStorage) GetProducts(ctx context.Context, limit int, offset int) ([]models.Product, error) {
	return f.getProductsFunc(ctx, limit, offset)
}

func (f *fakeStorage) DeleteProduct(ctx context.Context, id int64) error {
	return f.deleteProductFunc(ctx, id)
}

func (f *fakeStorage) UpdateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	return f.updateProductFunc(ctx, product)
}

func (f *fakeStorage) CreateTemplate(ctx context.Context, template models.Template) (models.Template, error) {
	return f.createTemplateFunc(ctx, template)
}

func (f *fakeStorage) GetTemplate(ctx context.Context, id int64) (models.Template, error) {
	return f.getTemplateFunc(ctx, id)
}

func (f *fakeStorage) GetTemplates(ctx context.Context, limit int, offset int) ([]models.Template, error) {
	return f.getTemplatesFunc(ctx, limit, offset)
}

func (f *fakeStorage) DeleteTemplate(ctx context.Context, id int64) error {
	return f.deleteTemplateFunc(ctx, id)
}

func (f *fakeStorage) UpdateTemplate(ctx context.Context, template models.Template) (models.Template, error) {
	return f.updateTemplateFunc(ctx, template)
}

func (f *fakeStorage) GetUserByTelegramID(ctx context.Context, userTelegramID int64) (*models.User, error) {
	return f.getUserByTelegramIDFunc(ctx, userTelegramID)
}

var _ Storage = (*fakeStorage)(nil)
