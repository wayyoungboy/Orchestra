package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/orchestra/backend/internal/models"
)

// UserRepository handles user database operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*models.User, error)
}

type sqlUserRepo struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &sqlUserRepo{db: db}
}

func (r *sqlUserRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, created_at, last_login_at)
		VALUES (?, ?, ?, ?, ?)
	`

	var lastLogin interface{}
	if user.LastLoginAt != nil {
		lastLogin = user.LastLoginAt.Unix()
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.CreatedAt.Unix(),
		lastLogin,
	)

	return err
}

func (r *sqlUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, last_login_at
		FROM users WHERE id = ?
	`

	user := &models.User{}
	var createdAt int64
	var lastLogin sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&createdAt,
		&lastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	if lastLogin.Valid {
		t := time.Unix(lastLogin.Int64, 0)
		user.LastLoginAt = &t
	}

	return user, nil
}

func (r *sqlUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, last_login_at
		FROM users WHERE username = ?
	`

	user := &models.User{}
	var createdAt int64
	var lastLogin sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&createdAt,
		&lastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	if lastLogin.Valid {
		t := time.Unix(lastLogin.Int64, 0)
		user.LastLoginAt = &t
	}

	return user, nil
}

func (r *sqlUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	return err
}

func (r *sqlUserRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqlUserRepo) List(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, last_login_at
		FROM users ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		var createdAt int64
		var lastLogin sql.NullInt64

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&createdAt,
			&lastLogin,
		); err != nil {
			return nil, err
		}

		user.CreatedAt = time.Unix(createdAt, 0)
		if lastLogin.Valid {
			t := time.Unix(lastLogin.Int64, 0)
			user.LastLoginAt = &t
		}
		users = append(users, user)
	}

	return users, nil
}