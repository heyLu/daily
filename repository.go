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
