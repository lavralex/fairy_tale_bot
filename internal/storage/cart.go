package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

var ErrOneOptionPerField = errors.New("can be only one option per field in cart item")

func (s *Storage) CreateCartItem(ctx context.Context, item models.CartItem) (models.CartItem, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.CartItem{}, fmt.Errorf("CreateCartItem begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(
		ctx,
		"INSERT INTO cart_items (user_id, product_id, comment) "+
			"VALUES ($1, $2, $3) RETURNING id, user_id, product_id, comment, created_at;",
		item.UserID, item.ProductID, item.Comment,
	)
	var returnedCartItem models.CartItem
	err = row.Scan(
		&returnedCartItem.ID,
		&returnedCartItem.UserID,
		&returnedCartItem.ProductID,
		&returnedCartItem.Comment,
		&returnedCartItem.CreatedAt,
	)
	var pgErr *pgconn.PgError
	// ошибка нарушения ограничения внешнего ключа, если поля к которому привязывается опция не существует foreign_key_violation
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "cart_items_user_id_fkey" {
		return models.CartItem{}, ErrUserNotFound
	}
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "cart_items_product_id_fkey" {
		return models.CartItem{}, ErrProductNotFound
	}
	if err != nil {
		return models.CartItem{}, fmt.Errorf("CreateCartItem cart item scan: %w", err)
	}
	returnedCartItemOptions := make([]models.CartItemOption, 0, len(item.Options))
	for _, option := range item.Options {
		row = tx.QueryRow(
			ctx,
			"INSERT INTO cart_item_options (cart_item_id, field_id, option_id) "+
				"VALUES ($1, $2, $3) RETURNING id, cart_item_id, field_id, option_id, created_at;",
			returnedCartItem.ID, option.FieldID, option.OptionID,
		)
		var returnedCartItemOption models.CartItemOption
		err = row.Scan(
			&returnedCartItemOption.ID,
			&returnedCartItemOption.CartItemID,
			&returnedCartItemOption.FieldID,
			&returnedCartItemOption.OptionID,
			&returnedCartItemOption.CreatedAt,
		)
		// ошибка нарушения ограничения внешнего ключа, если поля к которому привязывается опция не существует foreign_key_violation
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "cart_item_options_field_id_fkey" {
			return models.CartItem{}, ErrFieldNotFound
		}
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "cart_item_options_option_id_fkey" {
			return models.CartItem{}, ErrOptionNotFound
		}
		// ошибка нарушения ограничения уникального сочетания unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.CartItem{}, ErrOneOptionPerField
		}
		if err != nil {
			return models.CartItem{}, fmt.Errorf("CreateCartItem cart item option scan: %w", err)
		}
		returnedCartItemOptions = append(returnedCartItemOptions, returnedCartItemOption)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return models.CartItem{}, fmt.Errorf("CreateCartItem commit: %w", err)
	}
	returnedCartItem.Options = returnedCartItemOptions
	return returnedCartItem, nil
}

func (s *Storage) GetCart(ctx context.Context, userID int64) ([]models.CartItem, error) {
	rows, err := s.db.Query(
		ctx,
		"SELECT ci.id, ci.user_id, ci.product_id, p.price, p.name, ci.comment, ci.created_at FROM cart_items As ci "+
			"JOIN products AS p ON  p.id = ci.product_id "+
			"WHERE user_id = $1;", userID,
	)
	if err != nil {
		return nil, fmt.Errorf("GetCart cart items query: %w", err)
	}
	defer rows.Close()
	cart := make([]models.CartItem, 0)
	var itemsIDs = make([]int64, 0)
	var itemsMap = make(map[int64]*models.CartItem)
	for rows.Next() {
		var item models.CartItem
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.Price,
			&item.Name,
			&item.Comment,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetCart cart items scan: %w", err)
		}
		cart = append(cart, item)
		itemsIDs = append(itemsIDs, item.ID)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetCart cart items rows: %w", err)
	}
	for i := range cart {
		itemsMap[cart[i].ID] = &cart[i]
	}
	rows, err = s.db.Query(
		ctx,
		"SELECT cio.id, cio.cart_item_id, cio.field_id, f.name, cio.option_id, fo.name, cio.created_at FROM cart_item_options AS cio "+
			"JOIN fields AS f ON f.id = cio.field_id "+
			"JOIN field_options AS fo ON fo.id = cio.option_id "+
			"WHERE cio.cart_item_id = ANY($1);",
		itemsIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("GetCart options query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var option models.CartItemOption
		err := rows.Scan(
			&option.ID,
			&option.CartItemID,
			&option.FieldID,
			&option.FieldName,
			&option.OptionID,
			&option.Name,
			&option.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetCart options Scan: %w", err)
		}
		itemsMap[option.CartItemID].Options = append(itemsMap[option.CartItemID].Options, option)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetCart options rows: %w", err)
	}
	return cart, nil
}

var ErrCartItemNotFound = errors.New("cart item not found")

func (s *Storage) DeleteCartItem(ctx context.Context, id int64) error {
	ct, err := s.db.Exec(ctx, "DELETE FROM cart_items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteCartItem: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}
