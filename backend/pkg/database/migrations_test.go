package database

import (
	"io/fs"
	"testing"
)

func TestGetMigrationsFS(t *testing.T) {
	subFS, err := GetMigrationsFS("v0")
	if err != nil {
		t.Fatalf("expected no error getting migrations FS, got %v", err)
	}

	entries, err := fs.ReadDir(subFS, ".")
	if err != nil {
		t.Fatalf("expected no error reading subFS, got %v", err)
	}

	if len(entries) < 8 {
		t.Errorf("expected at least 8 migration files in v0, found %d", len(entries))
	}
}
