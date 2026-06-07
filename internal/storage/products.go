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
) (models.Product, error) {
	images := product.Images
	if imagesCount := len(images); imagesCount < 1 || imagesCount > 5 {
		return models.Product{}, fmt.Errorf("the incorrect number of images %d required is between 1 and 5", imagesCount)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Product{}, fmt.Errorf("CreateProduct: %w", err)
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(
		ctx,
		"INSERT INTO products (name, description, price, template_id, tags, is_available, comment_enabled, template_enabled) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, name, description, price, template_id, tags, created_at, is_available, comment_enabled, template_enabled;",
		product.Name, product.Description, product.Price, product.TemplateID, product.Tags, product.IsAvailable, product.CommentEnabled, product.TemplateEnabled,
	)
	var returnedProduct models.Product
	err = row.Scan(
		&returnedProduct.ID,
		&returnedProduct.Name,
		&returnedProduct.Description,
		&returnedProduct.Price,
		&returnedProduct.TemplateID,
		&returnedProduct.Tags,
		&returnedProduct.CreatedAt,
		&returnedProduct.IsAvailable,
		&returnedProduct.CommentEnabled,
		&returnedProduct.TemplateEnabled,
	)
	if err != nil {
		return models.Product{}, fmt.Errorf("CreateProduct: %w", err)
	}
	returnedImages := make([]models.ProductImage, 0, len(images))
	for _, img := range images {
		row = tx.QueryRow(
			ctx,
			"INSERT INTO product_images (product_id, src, alt, sort_order) "+
				"VALUES ($1, $2, $3, $4) RETURNING id, product_id, src, alt, sort_order, created_at;",
			returnedProduct.ID, img.Src, img.Alt, img.SortOrder,
		)
		var returnedImage models.ProductImage
		err = row.Scan(
			&returnedImage.ID,
			&returnedImage.ProductID,
			&returnedImage.Src,
			&returnedImage.Alt,
			&returnedImage.SortOrder,
			&returnedImage.CreatedAt,
		)
		if err != nil {
			return models.Product{}, fmt.Errorf("CreateProduct: %w", err)
		}
		returnedImages = append(returnedImages, returnedImage)
	}
	returnedProduct.Images = returnedImages
	err = tx.Commit(ctx)
	if err != nil {
		return models.Product{}, fmt.Errorf("CreateProduct: %w", err)
	}
	return returnedProduct, nil
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
	product.Images = make([]models.ProductImage, 0, 5)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetProduct product query: %w", err)
	}
	rows, err := s.db.Query(
		ctx,
		"SELECT id, product_id, src, alt, sort_order, created_at FROM product_images "+
			"WHERE product_id = $1",
		product.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("GetProduct image query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var image models.ProductImage
		err := rows.Scan(
			&image.ID,
			&image.ProductID,
			&image.Src,
			&image.Alt,
			&image.SortOrder,
			&image.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetProduct image Scan: %w", err)
		}
		product.Images = append(product.Images, image)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetProduct rows: %w", err)
	}
	return &product, nil
}
