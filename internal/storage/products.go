package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

var ErrProductNotFound = errors.New("product not found")
var ErrIncorrectNumberOfImages = errors.New("the incorrect number of images required is between 1 and 5")

func (s *Storage) CreateProduct(
	ctx context.Context,
	product models.Product,
) (models.Product, error) {
	images := product.Images
	if len(images) < 1 || len(images) > 5 {
		return models.Product{}, ErrIncorrectNumberOfImages
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Product{}, fmt.Errorf("CreateProduct: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
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

func (s *Storage) GetProduct(ctx context.Context, id int64) (models.Product, error) {
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
		return models.Product{}, ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("GetProduct product query: %w", err)
	}
	rows, err := s.db.Query(
		ctx,
		"SELECT id, product_id, src, alt, sort_order, created_at FROM product_images "+
			"WHERE product_id = $1",
		product.ID,
	)
	if err != nil {
		return models.Product{}, fmt.Errorf("GetProduct image query: %w", err)
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
			return models.Product{}, fmt.Errorf("GetProduct image Scan: %w", err)
		}
		product.Images = append(product.Images, image)
	}
	if err = rows.Err(); err != nil {
		return models.Product{}, fmt.Errorf("GetProduct rows: %w", err)
	}
	return product, nil
}

func (s *Storage) GetProducts(ctx context.Context, limit int, offset int) ([]models.Product, error) {
	rows, err := s.db.Query(
		ctx,
		"SELECT id, name, description, price, template_id, tags, created_at, is_available, comment_enabled, template_enabled FROM products "+
			"LIMIT $1 OFFSET $2", limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("GetProducts query: %w", err)
	}
	defer rows.Close()
	var products = make([]models.Product, 0, limit)
	var productsIDs = make([]int64, 0, limit)
	var productsMap = make(map[int64]*models.Product)
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("GetProducts product Scan: %w", err)
		}
		products = append(products, product)
		productsIDs = append(productsIDs, product.ID)
		productsMap[product.ID] = &products[len(products)-1]
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetProducts products rows: %w", err)
	}

	rows, err = s.db.Query(
		ctx,
		"SELECT id, product_id, src, alt, sort_order, created_at FROM product_images "+
			"WHERE product_id = ANY($1);",
		productsIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("GetProducts image query: %w", err)
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
			return nil, fmt.Errorf("GetProducts image Scan: %w", err)
		}
		productsMap[image.ProductID].Images = append(productsMap[image.ProductID].Images, image)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetProducts image rows: %w", err)
	}
	return products, nil
}

func (s *Storage) DeleteProduct(ctx context.Context, id int64) error {
	ct, err := s.db.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteProduct: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (s *Storage) UpdateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	images := product.Images
	if len(images) < 1 || len(images) > 5 {
		return models.Product{}, ErrIncorrectNumberOfImages
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Product{}, fmt.Errorf("UpdateProduct: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	row := tx.QueryRow(
		ctx,
		"UPDATE products SET name = $1, description = $2, price = $3, template_id = $4, tags = $5, is_available = $6, comment_enabled = $7, template_enabled = $8  WHERE id = $9 "+
			"RETURNING id, name, description, price, template_id, tags, created_at, is_available, comment_enabled, template_enabled",
		product.Name, product.Description, product.Price, product.TemplateID, product.Tags, product.IsAvailable, product.CommentEnabled, product.TemplateEnabled, product.ID,
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
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Product{}, ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("UpdateProduct products exec: %w", err)
	}
	_, err = tx.Exec(
		ctx,
		"DELETE FROM product_images WHERE product_id = $1",
		product.ID,
	)
	if err != nil {
		return models.Product{}, fmt.Errorf("UpdateProduct images delete: %w", err)
	}
	for _, img := range images {
		row := tx.QueryRow(
			ctx,
			"INSERT INTO product_images (product_id, src, alt, sort_order) "+
				"VALUES ($1, $2, $3, $4) RETURNING id, product_id, src, alt, sort_order, created_at;",
			product.ID, img.Src, img.Alt, img.SortOrder,
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
			return models.Product{}, fmt.Errorf("UpdateProduct image update: %w", err)
		}
		returnedProduct.Images = append(returnedProduct.Images, returnedImage)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return models.Product{}, fmt.Errorf("UpdateProduct commit: %w", err)
	}

	return returnedProduct, nil
}
