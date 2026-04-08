package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/orchestra/backend/internal/models"
)

type sqlAPIKeyRepo struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) APIKeyRepository {
	return &sqlAPIKeyRepo{db: db}
}

func (r *sqlAPIKeyRepo) Create(ctx context.Context, key *models.APIKey) error {
	query := `
		INSERT INTO api_keys (id, provider, encrypted_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		key.ID, key.Provider, key.EncryptedKey, key.CreatedAt.Unix(), key.UpdatedAt.Unix(),
	)
	return err
}

func (r *sqlAPIKeyRepo) GetByID(ctx context.Context, id string) (*models.APIKey, error) {
	query := `
		SELECT id, provider, encrypted_key, created_at, updated_at
		FROM api_keys WHERE id = ?
	`
	key := &models.APIKey{}
	var createdAt, updatedAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&key.ID, &key.Provider, &key.EncryptedKey, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	key.CreatedAt = time.Unix(createdAt, 0)
	key.UpdatedAt = time.Unix(updatedAt, 0)
	return key, nil
}

func (r *sqlAPIKeyRepo) GetByProvider(ctx context.Context, provider models.APIKeyProvider) (*models.APIKey, error) {
	query := `
		SELECT id, provider, encrypted_key, created_at, updated_at
		FROM api_keys WHERE provider = ?
	`
	key := &models.APIKey{}
	var createdAt, updatedAt int64
	err := r.db.QueryRowContext(ctx, query, provider).Scan(
		&key.ID, &key.Provider, &key.EncryptedKey, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	key.CreatedAt = time.Unix(createdAt, 0)
	key.UpdatedAt = time.Unix(updatedAt, 0)
	return key, nil
}

func (r *sqlAPIKeyRepo) List(ctx context.Context) ([]*models.APIKey, error) {
	query := `
		SELECT id, provider, encrypted_key, created_at, updated_at
		FROM api_keys ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]*models.APIKey, 0)
	for rows.Next() {
		key := &models.APIKey{}
		var createdAt, updatedAt int64
		if err := rows.Scan(
			&key.ID, &key.Provider, &key.EncryptedKey, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		key.CreatedAt = time.Unix(createdAt, 0)
		key.UpdatedAt = time.Unix(updatedAt, 0)
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *sqlAPIKeyRepo) Update(ctx context.Context, key *models.APIKey) error {
	query := `
		UPDATE api_keys SET encrypted_key = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		key.EncryptedKey, key.UpdatedAt.Unix(), key.ID,
	)
	return err
}

func (r *sqlAPIKeyRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqlAPIKeyRepo) DeleteByProvider(ctx context.Context, provider models.APIKeyProvider) error {
	query := `DELETE FROM api_keys WHERE provider = ?`
	_, err := r.db.ExecContext(ctx, query, provider)
	return err
}