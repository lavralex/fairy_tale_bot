package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/lavralex/fairy_tale_bot/internal/models"
)

var ErrIncorrectNumberOfFields = errors.New("the incorrect number of fields required is more then 0")

func (s *Storage) CreateTemplate(
	ctx context.Context,
	template models.Template,
) (models.Template, error) {
	if len(template.Fields) < 1 {
		return models.Template{}, ErrIncorrectNumberOfFields
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate: %w", err)
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(
		ctx,
		"INSERT INTO templates (name) "+
			"VALUES ($1) RETURNING id, name, created_at;",
		template.Name,
	)
	var returnedTemplate models.Template
	err = row.Scan(
		&returnedTemplate.ID,
		&returnedTemplate.Name,
		&returnedTemplate.CreatedAt,
	)
	if err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate: %w", err)
	}
	for order, field := range template.Fields {
		_, err := tx.Exec(
			ctx,
			"INSERT INTO template_fields (template_id, field_id, sort_order) "+
				"VALUES ($1, $2, $3);",
			returnedTemplate.ID, field.Field.ID, order,
		)
		if err != nil {
			return models.Template{}, fmt.Errorf("CreateTemplate: %w", err)
		}
	}
	rows, err := tx.Query(
		ctx,
		"SELECT f.id, f.name, f.is_active, f.created_at, tf.sort_order "+
			"FROM template_fields AS tf "+
			"JOIN fields AS f ON f.id = tf.field_id "+
			"WHERE tf.template_id = $1 "+
			"ORDER BY tf.sort_order",
		returnedTemplate.ID,
	)
	if err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate fields query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var field models.Field
		var order int
		err = rows.Scan(
			&field.ID,
			&field.Name,
			&field.IsActive,
			&field.CreatedAt,
			&order,
		)
		if err != nil {
			return models.Template{}, fmt.Errorf("CreateTemplate field scan: %w", err)
		}
		returnedTemplate.Fields = append(
			returnedTemplate.Fields,
			models.TemplateField{
				Field:     field,
				SortOrder: order,
			},
		)
	}
	if err = rows.Err(); err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate fields rows: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate commit: %w", err)
	}
	return returnedTemplate, nil
}
