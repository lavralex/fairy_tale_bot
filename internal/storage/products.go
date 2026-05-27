package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

var ErrProductNotFound = errors.New("product not found")

func (s *Storage) CreateProduct(
	ctx context.Context,
	product models.Product,
	images []models.ProductImage,
) error {
	if imagesCount := len(images); imagesCount < 1 || imagesCount > 5 {
		return fmt.Errorf("the incorrect number of images %d required is between 1 and 5", imagesCount)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("CreateProduct: %w", err)
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(
		ctx,
		"INSERT INTO products (name, description, price, template_id, tags, is_available, comment_enabled, template_enabled) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;",
		product.Name, product.Description, product.Price, product.TemplateID, product.Tags, product.IsAvailable, product.CommentEnabled, product.TemplateEnabled,
	)
	var productID int64
	err = row.Scan(&productID)
	if err != nil {
		return fmt.Errorf("CreateProduct: %w", err)
	}
	for _, img := range images {
		_, err = tx.Exec(
			ctx,
			"INSERT INTO product_images (product_id, src, alt, sort_order) "+
				"VALUES ($1, $2, $3, $4);",
			productID, img.Src, img.Alt, img.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("CreateProduct: %w", err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("CreateProduct: %w", err)
	}
	return nil
}

func (s *Storage) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	var product models.Product
	row := s.db.QueryRow(
		ctx,
		"SELECT id, name, description, price, template_id, tags, created_at, is_available, comment_enabled, template_enabled FROM products "+
			"WHERE id = $1", id,
	)
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.TemplateID,
		&product.Tags,
		&product.CreatedAt,
		&product.IsAvailable,
		&product.CommentEnabled,
		&product.TemplateEnabled,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetProduct: %w", err)
	}
	return &product, nil
}
