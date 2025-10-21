package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS json_storage (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		creator_key TEXT NOT NULL DEFAULT 'guest'
	);

	CREATE INDEX IF NOT EXISTS idx_json_storage_expires_at ON json_storage(expires_at);
	CREATE INDEX IF NOT EXISTS idx_json_storage_id_expires ON json_storage(id, expires_at);

	CREATE TABLE IF NOT EXISTS api_keys (
		key TEXT PRIMARY KEY,
		description TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_admin BOOLEAN NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) CreateJSON(ctx context.Context, id, data string, expiresAt time.Time, creatorKey string) error {
	query := `INSERT INTO json_storage (id, data, expires_at, creator_key) VALUES (?, ?, ?, ?)`
	_, err := db.conn.ExecContext(ctx, query, id, data, expiresAt, creatorKey)
	return err
}

func (db *DB) GetJSON(ctx context.Context, id string) (string, error) {
	query := `SELECT data FROM json_storage WHERE id = ? AND expires_at > ?`
	var data string
	err := db.conn.QueryRowContext(ctx, query, id, time.Now()).Scan(&data)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("json not found")
	}
	if err != nil {
		return "", err
	}
	return data, nil
}

func (db *DB) DeleteExpiredJSON(ctx context.Context) (int64, error) {
	query := `DELETE FROM json_storage WHERE expires_at < ?`
	result, err := db.conn.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *DB) CreateApiKey(ctx context.Context, key, description string, isAdmin bool) error {
	query := `INSERT INTO api_keys (key, description, is_admin, created_at) VALUES (?, ?, ?, ?)`
	_, err := db.conn.ExecContext(ctx, query, key, description, isAdmin, time.Now())
	return err
}

func (db *DB) GetApiKey(ctx context.Context, key string) (isAdmin bool, createdAt time.Time, err error) {
	query := `SELECT is_admin, created_at FROM api_keys WHERE key = ?`
	err = db.conn.QueryRowContext(ctx, query, key).Scan(&isAdmin, &createdAt)
	if err == sql.ErrNoRows {
		return false, time.Time{}, fmt.Errorf("api key not found")
	}
	return isAdmin, createdAt, err
}

func (db *DB) DeleteApiKey(ctx context.Context, key string) error {
	query := `DELETE FROM api_keys WHERE key = ?`
	result, err := db.conn.ExecContext(ctx, query, key)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}
