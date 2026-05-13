package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lavralex/fairy_tale_bot/internal/models"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New err: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pool.Ping err: %w", err)
	}
	return &Storage{db: pool}, nil
}

func (s *Storage) CreateUser(ctx context.Context, user models.User) error {
	_, err := s.db.Exec(
		ctx,
		"INSERT INTO users (telegram_id, username, phone, email) "+
			"VALUES ($1, $2, $3, $4)", user.TelegramID, user.Username, user.Phone, user.Email,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil
			}
			return fmt.Errorf("create user DB err: %w", err)
		}
		return fmt.Errorf("create user err: %w", err)
	}
	return nil
}
