package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

func (s *Storage) CreateUser(ctx context.Context, user models.User) error {
	_, err := s.db.Exec(
		ctx,
		"INSERT INTO users (telegram_id, username, phone, email) "+
			"VALUES ($1, $2, $3, $4);", user.TelegramID, user.Username, user.Phone, user.Email,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // 23505 - unique violation
				return nil
			}
			return fmt.Errorf("CreateUser DB: %w", err)
		}
		return fmt.Errorf("CreateUser: %w", err)
	}
	return nil
}

func (s *Storage) GetUserByTelegramID(ctx context.Context, userTelegramID int64) (*models.User, error) {
	var user models.User
	row := s.db.QueryRow(
		ctx,
		"SELECT id, telegram_id, username, phone, email, created_at, role, is_banned FROM users "+
			"WHERE telegram_id = $1;", userTelegramID,
	)
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Phone,
		&user.Email,
		&user.CreatedAt,
		&user.Role,
		&user.IsBanned,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetUserByTelegramID: %w", err)
	}
	return &user, nil
}
