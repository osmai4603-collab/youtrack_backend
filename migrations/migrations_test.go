package migrations

import (
	"io"
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func TestEmbeddedMigrationSixIsReadable(t *testing.T) {
	source, err := iofs.New(FS, ".")
	if err != nil {
		t.Fatalf("create embedded migration source: %v", err)
	}
	defer source.Close()

	migration, identifier, err := source.ReadUp(6)
	if err != nil {
		t.Fatalf("read migration version 6: %v", err)
	}
	defer migration.Close()

	if identifier == "" {
		t.Fatal("migration version 6 has no identifier")
	}
	if _, err := io.ReadAll(migration); err != nil {
		t.Fatalf("read migration version 6 contents: %v", err)
	}
}
