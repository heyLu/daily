package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Repository interface {
	Create(ctx context.Context, entry Entry) (id string, err error)
	Get(ctx context.Context, id string) (*Entry, error)
}

func NewRepository(dbFileName string) (Repository, error) {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		return nil, fmt.Errorf("could not open db in %q: %s", dbFileName, err)
	}

	return &repository{db: db}, nil
}

type repository struct {
	db *sql.DB
}

func (r *repository) Create(ctx context.Context, entry Entry) (id string, err error) {
	id, err = generateID()
	if err != nil {
		return "", fmt.Errorf("could not generate id: %s", err)
	}

	dataJSON, err := json.Marshal(entry.Data)
	if err != nil {
		return "", fmt.Errorf("could not serialize additional data: %s", err)
	}

	_, err = r.db.ExecContext(ctx, "INSERT INTO entries (id, date, type, note, value, data) VALUES (?, ?, ?, ?, ?, ?)",
		id, entry.Date, entry.Type, entry.Note, entry.Value, dataJSON)
	if err != nil {
		return "", fmt.Errorf("could not store entry: %s", err)
	}

	return id, nil
}

func generateID() (string, error) {
	buf := make([]byte, 12)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}


func (r *repository) Get(ctx context.Context, id string) (*Entry, error) {
	var entry Entry
	var rawData []byte
	row := r.db.QueryRowContext(ctx, "SELECT id, date, type, note, value, data FROM entries WHERE id = ?", id)
	err := row.Scan(&entry.ID, &entry.Date, &entry.Type, &entry.Note, &entry.Value, &rawData)
	if err != nil {
		if err == sql.ErrNoRows {
			// not found
			return nil, nil
		}

		return nil, fmt.Errorf("could not get entry with id %q: %s", id, err)
	}

	if len(rawData) > 0 {
		var data map[string]interface{}
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return nil, fmt.Errorf("additional data %q was invalid: %s", string(rawData), err)
		}
		entry.Data = data
	}

	return &entry, nil
}
