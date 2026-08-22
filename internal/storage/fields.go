package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

var ErrIncorrectNumberOfOptions = errors.New("the incorrect number of options required is more then 0")

func (s *Storage) CreateField(
	ctx context.Context,
	field models.Field,
) (models.Field, error) {
	if len(field.Options) < 1 {
		return models.Field{}, ErrIncorrectNumberOfOptions
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Field{}, fmt.Errorf("CreateField: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(
		ctx,
		"INSERT INTO fields (name, is_active, created_at) "+
			"VALUES ($1, $2, $3) RETURNING id, name, is_active, created_at;",
		field.Name, field.IsActive, field.CreatedAt,
	)
	var returnedField models.Field
	err = row.Scan(
		&returnedField.ID,
		&returnedField.Name,
		&returnedField.IsActive,
		&returnedField.CreatedAt,
	)
	if err != nil {
		return models.Field{}, fmt.Errorf("CreateField: %w", err)
	}
	returnedOptions := make([]models.FieldOption, 0, len(field.Options))

	for _, option := range field.Options {
		row = tx.QueryRow(
			ctx,
			"INSERT INTO field_options (field_id, name, sort_order, created_at, is_active) "+
				"VALUES ($1, $2, $3, $4, $5) RETURNING id, field_id, name, sort_order, created_at, is_active;",
			returnedField.ID, option.Name, option.SortOrder, option.CreatedAt, option.IsActive,
		)
		var returnedOption models.FieldOption
		err = row.Scan(
			&returnedOption.ID,
			&returnedOption.FieldID,
			&returnedOption.Name,
			&returnedOption.SortOrder,
			&returnedOption.CreatedAt,
			&returnedOption.IsActive,
		)
		if err != nil {
			return models.Field{}, fmt.Errorf("CreateField: %w", err)
		}
		returnedOptions = append(returnedOptions, returnedOption)
	}
	returnedField.Options = returnedOptions
	err = tx.Commit(ctx)
	if err != nil {
		return models.Field{}, fmt.Errorf("CreateField: %w", err)
	}
	return returnedField, nil
}

var ErrFieldNotFound = errors.New("field not found")

func (s *Storage) GetField(ctx context.Context, id int64) (models.Field, error) {
	var field models.Field
	row := s.db.QueryRow(
		ctx,
		"SELECT id, name, is_active, created_at FROM fields "+
			"WHERE id = $1", id,
	)
	err := row.Scan(
		&field.ID,
		&field.Name,
		&field.IsActive,
		&field.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Field{}, ErrFieldNotFound
	}
	if err != nil {
		return models.Field{}, fmt.Errorf("GetField field query: %w", err)
	}

	rows, err := s.db.Query(
		ctx,
		"SELECT id, field_id, name, is_active, sort_order, created_at FROM field_options "+
			"WHERE field_id = $1",
		field.ID,
	)
	if err != nil {
		return models.Field{}, fmt.Errorf("GetField option query: %w", err)
	}
	defer rows.Close()
	field.Options = make([]models.FieldOption, 0)
	for rows.Next() {
		var option models.FieldOption
		err := rows.Scan(
			&option.ID,
			&option.FieldID,
			&option.Name,
			&option.IsActive,
			&option.SortOrder,
			&option.CreatedAt,
		)
		if err != nil {
			return models.Field{}, fmt.Errorf("GetField option Scan: %w", err)
		}
		field.Options = append(field.Options, option)
	}
	if err = rows.Err(); err != nil {
		return models.Field{}, fmt.Errorf("GetField rows: %w", err)
	}
	return field, nil
}

func (s *Storage) GetFields(ctx context.Context) ([]models.Field, error) {
	rows, err := s.db.Query(
		ctx,
		"SELECT id, name, is_active, created_at FROM fields",
	)
	if err != nil {
		return nil, fmt.Errorf("GetFields query: %w", err)
	}
	defer rows.Close()
	var fields = make([]models.Field, 0)
	var fieldsIDs = make([]int64, 0)
	var fieldsMap = make(map[int64]*models.Field)
	for rows.Next() {
		var field models.Field
		err := rows.Scan(
			&field.ID,
			&field.Name,
			&field.IsActive,
			&field.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetFields field Scan: %w", err)
		}
		fields = append(fields, field)
		fieldsIDs = append(fieldsIDs, field.ID)
		fieldsMap[field.ID] = &fields[len(fields)-1]
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetFields fields rows: %w", err)
	}
	rows, err = s.db.Query(
		ctx,
		"SELECT id, field_id, name, sort_order, created_at, is_active FROM field_options "+
			"WHERE field_id = ANY($1);",
		fieldsIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("GetFields option query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var option models.FieldOption
		err := rows.Scan(
			&option.ID,
			&option.FieldID,
			&option.Name,
			&option.SortOrder,
			&option.CreatedAt,
			&option.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("GetFields options Scan: %w", err)
		}
		fieldsMap[option.FieldID].Options = append(fieldsMap[option.FieldID].Options, option)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetFields options rows: %w", err)
	}
	return fields, nil
}

func (s *Storage) DeleteField(ctx context.Context, id int64) error {
	ct, err := s.db.Exec(ctx, "DELETE FROM fields WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteField: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrFieldNotFound
	}
	return nil
}

func (s *Storage) UpdateField(ctx context.Context, field models.Field) (models.Field, error) {
	row := s.db.QueryRow(
		ctx,
		"UPDATE fields SET name = $1, is_active = $2 WHERE id = $3 "+
			"RETURNING id, name, is_active, created_at",
		field.Name, field.IsActive, field.ID,
	)
	var returnedField models.Field
	err := row.Scan(
		&returnedField.ID,
		&returnedField.Name,
		&returnedField.IsActive,
		&returnedField.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Field{}, ErrFieldNotFound
	}
	if err != nil {
		return models.Field{}, fmt.Errorf("UpdateField field QueryRow: %w", err)
	}
	rows, err := s.db.Query(
		ctx,
		"SELECT id, field_id, name, is_active, sort_order, created_at FROM field_options "+
			"WHERE field_id = $1",
		field.ID,
	)
	if err != nil {
		return models.Field{}, fmt.Errorf("UpdateField option query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var option models.FieldOption
		err := rows.Scan(
			&option.ID,
			&option.FieldID,
			&option.Name,
			&option.IsActive,
			&option.SortOrder,
			&option.CreatedAt,
		)
		if err != nil {
			return models.Field{}, fmt.Errorf("UpdateField option Scan: %w", err)
		}
		returnedField.Options = append(returnedField.Options, option)
	}
	if err = rows.Err(); err != nil {
		return models.Field{}, fmt.Errorf("UpdateField option rows: %w", err)
	}
	return returnedField, nil
}
