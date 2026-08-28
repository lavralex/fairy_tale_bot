package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

func (s *Storage) CreateOption(ctx context.Context, option models.FieldOption) (models.FieldOption, error) {
	row := s.db.QueryRow(
		ctx,
		"INSERT INTO field_options (field_id, name, sort_order, created_at, is_active) "+
			"VALUES ($1, $2, $3, $4, $5) RETURNING id, field_id, name, sort_order, created_at, is_active;",
		option.FieldID, option.Name, option.SortOrder, option.CreatedAt, option.IsActive,
	)
	var returnedOption models.FieldOption
	err := row.Scan(
		&returnedOption.ID,
		&returnedOption.FieldID,
		&returnedOption.Name,
		&returnedOption.SortOrder,
		&returnedOption.CreatedAt,
		&returnedOption.IsActive,
	)
	var pgErr *pgconn.PgError
	// ошибка нарушения ограничения внешнего ключа, если поля к которому привязывается опция не существует
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return models.FieldOption{}, ErrFieldNotFound
	}
	if err != nil {
		return models.FieldOption{}, fmt.Errorf("CreateOption: %w", err)
	}
	return returnedOption, nil
}

var ErrOptionNotFound = errors.New("option not found")

func (s *Storage) DeleteOption(ctx context.Context, id int64) error {
	ct, err := s.db.Exec(ctx, "DELETE FROM field_options WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteOption: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrOptionNotFound
	}
	return nil
}

func (s *Storage) UpdateOption(ctx context.Context, option models.FieldOption) (models.FieldOption, error) {
	row := s.db.QueryRow(
		ctx,
		"UPDATE field_options SET name = $1, sort_order = $2, is_active = $3 WHERE id = $4 "+
			"RETURNING id, field_id, name, sort_order, is_active, created_at",
		option.Name, option.SortOrder, option.IsActive, option.ID,
	)
	var returnedOption models.FieldOption
	err := row.Scan(
		&returnedOption.ID,
		&returnedOption.FieldID,
		&returnedOption.Name,
		&returnedOption.SortOrder,
		&returnedOption.IsActive,
		&returnedOption.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.FieldOption{}, ErrOptionNotFound
	}
	if err != nil {
		return models.FieldOption{}, fmt.Errorf("UpdateOption QueryRow: %w", err)
	}
	return returnedOption, nil
}
