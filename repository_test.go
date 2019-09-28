package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestCreateEntry(t *testing.T) {
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("could not open repository: %s", err)
	}
	err = initSchema(context.Background(), repo.(*repository).db, "schema-init.sql")
	if err != nil {
		t.Fatalf("could not initialize schema: %s", err)
	}

	entry := Entry{
		Date: time.Now(),
		Type: "test",
		Note: "first ever entry!! 🎉",
	}
	id, err := repo.Create(context.Background(), entry)
	if err != nil {
		t.Fatalf("could not create entry: %s", err)
	}

	if id == "" {
		t.Fatal("id was empty")
	}

	t.Logf("created entry with id %q", id)
}

func initSchema(ctx context.Context, db *sql.DB, schemaFileName string) error {
	schemaSQL, err := ioutil.ReadFile(schemaFileName)
	if err != nil {
		return fmt.Errorf("could not read schema sql from %q: %s", schemaFileName, err)
	}

	statements := strings.Split(string(schemaSQL), ";")
	for _, stmt := range statements {
		_, err := db.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("could not execute %q: %s", stmt, err)
		}
	}

	return nil
}
