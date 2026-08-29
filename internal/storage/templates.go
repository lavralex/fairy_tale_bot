package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

func loadTemplateFields(ctx context.Context, tx pgx.Tx, id int64) ([]models.TemplateField, error) {
	rows, err := tx.Query(
		ctx,
		"SELECT f.id, f.name, f.is_active, f.created_at, tf.sort_order "+
			"FROM template_fields AS tf "+
			"JOIN fields AS f ON f.id = tf.field_id "+
			"WHERE tf.template_id = $1 "+
			"ORDER BY tf.sort_order",
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("fields query: %w", err)
	}
	defer rows.Close()
	fieldsIDs := make([]int64, 0)
	fieldsMap := make(map[int64]*models.Field)
	var templateFields []models.TemplateField
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
			return nil, fmt.Errorf("field scan: %w", err)
		}
		fieldsIDs = append(fieldsIDs, field.ID)
		templateFields = append(
			templateFields,
			models.TemplateField{
				Field:     field,
				SortOrder: order,
			},
		)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("fields rows: %w", err)
	}
	for i := range templateFields {
		fieldsMap[templateFields[i].Field.ID] = &templateFields[i].Field
	}
	rows, err = tx.Query(
		ctx,
		"SELECT id, field_id, name, is_active, sort_order, created_at FROM field_options "+
			"WHERE field_id = ANY($1) "+
			"ORDER BY sort_order",
		fieldsIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("option query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var option models.FieldOption
		err := rows.Scan(
			&option.ID,
			&option.Name,
			&option.FieldID,
			&option.IsActive,
			&option.SortOrder,
			&option.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("option Scan: %w", err)
		}
		fieldsMap[option.FieldID].Options = append(fieldsMap[option.FieldID].Options, option)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("option rows: %w", err)
	}
	return templateFields, nil
}

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
		return models.Template{}, fmt.Errorf("CreateTemplate begin: %w", err)
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
		return models.Template{}, fmt.Errorf("CreateTemplate template Scan: %w", err)
	}
	for _, field := range template.Fields {
		_, err := tx.Exec(
			ctx,
			"INSERT INTO template_fields (template_id, field_id, sort_order) "+
				"VALUES ($1, $2, $3);",
			returnedTemplate.ID, field.Field.ID, field.SortOrder,
		)
		if err != nil {
			return models.Template{}, fmt.Errorf("CreateTemplate template fields exec: %w", err)
		}
	}

	returnedTemplate.Fields, err = loadTemplateFields(ctx, tx, returnedTemplate.ID)
	if err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate load template fields: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return models.Template{}, fmt.Errorf("CreateTemplate commit: %w", err)
	}
	return returnedTemplate, nil
}

var ErrTemplateNotFound = errors.New("template not found")

func (s *Storage) GetTemplate(ctx context.Context, id int64) (models.Template, error) {
	row := s.db.QueryRow(
		ctx,
		"SELECT id, name, created_at FROM templates "+
			"WHERE id = $1;", id,
	)
	var template models.Template
	err := row.Scan(
		&template.ID,
		&template.Name,
		&template.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Template{}, ErrTemplateNotFound
	}
	if err != nil {
		return models.Template{}, fmt.Errorf("GetTemplate template query: %w", err)
	}
	rows, err := s.db.Query(
		ctx,
		"SELECT f.id, f.name, f.is_active, f.created_at, tf.sort_order "+
			"FROM template_fields AS tf "+
			"JOIN fields AS f ON f.id = tf.field_id "+
			"WHERE tf.template_id = $1 "+
			"ORDER BY tf.sort_order",
		template.ID,
	)
	if err != nil {
		return models.Template{}, fmt.Errorf("GetTemplate fields query: %w", err)
	}
	defer rows.Close()
	var fieldsIDs = make([]int64, 0)
	var fieldsMap = make(map[int64]*models.Field)
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
			return models.Template{}, fmt.Errorf("GetTemplate field scan: %w", err)
		}
		template.Fields = append(
			template.Fields,
			models.TemplateField{
				Field:     field,
				SortOrder: order,
			},
		)
		fieldsIDs = append(fieldsIDs, field.ID)
	}
	for i := range template.Fields {
		fieldsMap[template.Fields[i].Field.ID] = &template.Fields[i].Field
	}
	if err = rows.Err(); err != nil {
		return models.Template{}, fmt.Errorf("GetTemplate fields rows: %w", err)
	}
	rows, err = s.db.Query(
		ctx,
		"SELECT id, field_id, name, is_active, sort_order, created_at FROM field_options "+
			"WHERE field_id = ANY($1)",
		fieldsIDs,
	)
	if err != nil {
		return models.Template{}, fmt.Errorf("GetTemplate options query: %w", err)
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
			return models.Template{}, fmt.Errorf("GetTemplate options Scan: %w", err)
		}
		fieldsMap[option.FieldID].Options = append(fieldsMap[option.FieldID].Options, option)
	}
	if err = rows.Err(); err != nil {
		return models.Template{}, fmt.Errorf("GetTemplate options rows: %w", err)
	}
	return template, nil
}

func (s *Storage) GetTemplates(ctx context.Context, limit int, offset int) ([]models.Template, error) {
	rows, err := s.db.Query(
		ctx,
		"SELECT id, name, created_at FROM templates "+
			"LIMIT $1 OFFSET $2", limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTemplates query: %w", err)
	}
	defer rows.Close()
	var templates = make([]models.Template, 0)
	for rows.Next() {
		var template models.Template
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetTemplates Scan: %w", err)
		}
		templates = append(templates, template)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetTemplates rows: %w", err)
	}
	return templates, nil
}

func (s *Storage) DeleteTemplate(ctx context.Context, id int64) error {
	ct, err := s.db.Exec(ctx, "DELETE FROM templates WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("DeleteTemplate: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

func (s *Storage) UpdateTemplate(ctx context.Context, template models.Template) (models.Template, error) {
	if len(template.Fields) < 1 {
		return models.Template{}, ErrIncorrectNumberOfFields
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Template{}, fmt.Errorf("UpdateTemplate begin: %w", err)
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(
		ctx,
		"UPDATE templates SET name = $1 WHERE id = $2 "+
			"RETURNING id, name, created_at",
		template.Name, template.ID,
	)
	var returnedTemplate models.Template
	err = row.Scan(
		&returnedTemplate.ID,
		&returnedTemplate.Name,
		&returnedTemplate.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Template{}, ErrTemplateNotFound
	}
	if err != nil {
		return models.Template{}, fmt.Errorf("UpdateTemplate template scan: %w", err)
	}
	_, err = tx.Exec(
		ctx,
		"DELETE FROM template_fields WHERE template_id = $1", template.ID,
	)
	if err != nil {
		return models.Template{}, fmt.Errorf("UpdateTemplate fields delete: %w", err)
	}
	for _, field := range template.Fields {
		_, err := tx.Exec(
			ctx,
			"INSERT INTO template_fields (template_id, field_id, sort_order) "+
				"VALUES ($1, $2, $3);",
			returnedTemplate.ID, field.Field.ID, field.SortOrder,
		)
		if err != nil {
			return models.Template{}, fmt.Errorf("UpdateTemplate template fields exec: %w", err)
		}
	}
	returnedTemplate.Fields, err = loadTemplateFields(ctx, tx, returnedTemplate.ID)
	if err != nil {
		return models.Template{}, fmt.Errorf("UpdateTemplate load template fields: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return models.Template{}, fmt.Errorf("UpdateTemplate commit: %w", err)
	}
	return returnedTemplate, nil
}
