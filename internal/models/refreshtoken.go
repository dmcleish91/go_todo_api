package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	TokenID   int
	UserID    int
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenModel struct {
	DB *pgxpool.Pool
}

func (m *RefreshTokenModel) CreateRefreshToken(userID int, token string, expiresAt time.Time) error {
	query := `
	INSERT INTO refresh_tokens (user_id, token, expires_at)
	VALUES ($1, $2, $3)`

	_, err := m.DB.Exec(context.Background(), query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %v", err)
	}

	return nil
}

func (m *RefreshTokenModel) ValidateRefreshToken(token string) (*RefreshToken, error) {
	query := `
        SELECT token_id, user_id, token, expires_at, created_at
        FROM refresh_tokens
        WHERE token = $1 AND expires_at > NOW()`

	rt := &RefreshToken{}
	err := m.DB.QueryRow(context.Background(), query, token).Scan(
		&rt.TokenID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token: %v", err)
	}

	return rt, nil
}

func (m *RefreshTokenModel) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	_, err := m.DB.Exec(context.Background(), query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %v", err)
	}

	return nil
}
