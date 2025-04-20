// pkg/storage/storage.go
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store provides a simple JSON-backed key/value store in SQLite.
type Store struct {
	db *sql.DB
}

// New opens (or creates) the SQLite file at dbPath and ensures the state table exists.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}
	const createStmt = `
CREATE TABLE IF NOT EXISTS state (
  namespace TEXT NOT NULL,
  item_key  TEXT NOT NULL,
  data      TEXT NOT NULL,
  PRIMARY KEY (namespace, item_key)
);`
	if _, err := db.Exec(createStmt); err != nil {
		return nil, fmt.Errorf("failed to create state table: %w", err)
	}
	return &Store{db: db}, nil
}

// Save stores the JSON-serialized value under (namespace, key).
// On conflict it overwrites the existing data.
func (s *Store) Save(namespace, key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	const stmt = `
INSERT INTO state (namespace, item_key, data)
VALUES (?, ?, ?)
ON CONFLICT(namespace, item_key) DO UPDATE SET data=excluded.data;`
	if _, err := s.db.Exec(stmt, namespace, key, string(bytes)); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

// Load retrieves the JSON data for (namespace, key) and unmarshals it into value.
// If no row exists, it leaves value untouched (so callers can initialize it first).
func (s *Store) Load(namespace, key string, value interface{}) error {
	const query = `SELECT data FROM state WHERE namespace = ? AND item_key = ?;`
	row := s.db.QueryRow(query, namespace, key)

	var jsonData string
	if err := row.Scan(&jsonData); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to query state: %w", err)
	}
	if err := json.Unmarshal([]byte(jsonData), value); err != nil {
		return fmt.Errorf("failed to unmarshal state data: %w", err)
	}
	return nil
}