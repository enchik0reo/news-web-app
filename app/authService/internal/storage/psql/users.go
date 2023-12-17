package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"newsWebApp/app/authService/internal/models"
	"newsWebApp/app/authService/internal/storage"
)

func (s *Storage) SaveUser(ctx context.Context, userName string, email string, hashPass []byte) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (user_name, email, password_hash) VALUES ($1, $2, $3) RETURNING user_id")
	if err != nil {
		return 0, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, userName, email, hashPass)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("can't insert values: %w", err)
	}

	var id int64

	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("can't get last insert id: %w", err)
	}

	return id, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT user_id, user_name, email, password_hash FROM users WHERE email = $1")
	if err != nil {
		return nil, fmt.Errorf("can't prepare statement: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	if err := row.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("can't get user result: %w", err)
	}

	u := models.User{}

	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PassHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("can't get result: %w", err)
	}

	return &u, nil
}
