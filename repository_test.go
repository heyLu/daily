package main

import (
	"context"
	"testing"
	"time"
)

func TestCreateEntry(t *testing.T) {
	repo, err := NewRepository(":memory:", "./schema-init.sql")
	if err != nil {
		t.Fatalf("could not open repository: %s", err)
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
