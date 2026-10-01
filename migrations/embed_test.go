package migrations_test

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"youtrack_backend/migrations"
)

var filenameRegex = regexp.MustCompile(`^([0-9]{6})_(.*)\.(up|down)\.sql$`)

func TestMigrationSequenceAndInvariants(t *testing.T) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		t.Fatalf("failed to read embedded migrations FS: %v", err)
	}

	upFiles := make(map[int]string)
	downFiles := make(map[int]string)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasSuffix(name, ".go") {
			continue
		}

		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			t.Errorf("migration file %q has invalid prefix ('_' or '.')", name)
		}

		matches := filenameRegex.FindStringSubmatch(name)
		if matches == nil {
			t.Errorf("migration file %q does not match pattern ^([0-9]{6})_(.*).(up|down).sql$", name)
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			t.Errorf("invalid version number in filename %q: %v", name, err)
			continue
		}

		kind := matches[3]
		if kind == "up" {
			if existing, exists := upFiles[version]; exists {
				t.Errorf("duplicate UP migration version %06d: %q and %q", version, existing, name)
			}
			upFiles[version] = name
		} else if kind == "down" {
			if existing, exists := downFiles[version]; exists {
				t.Errorf("duplicate DOWN migration version %06d: %q and %q", version, existing, name)
			}
			downFiles[version] = name
		}
	}

	const expectedCount = 12

	if len(upFiles) != expectedCount {
		t.Errorf("expected exactly %d UP migration files, got %d", expectedCount, len(upFiles))
	}
	if len(downFiles) != expectedCount {
		t.Errorf("expected exactly %d DOWN migration files, got %d", expectedCount, len(downFiles))
	}

	for v := 1; v <= expectedCount; v++ {
		upName, hasUp := upFiles[v]
		if !hasUp {
			t.Errorf("missing UP migration for version %06d", v)
		}
		downName, hasDown := downFiles[v]
		if !hasDown {
			t.Errorf("missing DOWN migration for version %06d", v)
		}

		if hasUp && hasDown {
			upBase := strings.TrimSuffix(upName, ".up.sql")
			downBase := strings.TrimSuffix(downName, ".down.sql")
			if upBase != downBase {
				t.Errorf("mismatched UP and DOWN base names for version %06d: UP=%q DOWN=%q", v, upName, downName)
			}
		}
	}

	var versions []int
	for v := range upFiles {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	for i, v := range versions {
		expectedVersion := i + 1
		if v != expectedVersion {
			t.Errorf("non-contiguous migration version at index %d: expected %06d, got %06d", i, expectedVersion, v)
		}
	}

	// Content checks on UP files
	for _, name := range upFiles {
		contentBytes, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			t.Errorf("failed to read migration file %q: %v", name, err)
			continue
		}
		content := string(contentBytes)

		if strings.Contains(content, "\\echo") {
			t.Errorf("file %q contains forbidden psql command '\\echo'", name)
		}

		// Check for standalone BEGIN; or COMMIT; statements
		lines := strings.Split(content, "\n")
		for idx, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.EqualFold(trimmed, "BEGIN;") {
				t.Errorf("file %q line %d contains forbidden 'BEGIN;' statement", name, idx+1)
			}
			if strings.EqualFold(trimmed, "COMMIT;") {
				t.Errorf("file %q line %d contains forbidden 'COMMIT;' statement", name, idx+1)
			}
		}

		if strings.Contains(content, "schema_migrations") {
			t.Errorf("file %q contains forbidden reference to 'schema_migrations'", name)
		}
	}
}

func TestMigrationFileCount(t *testing.T) {
	files, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("failed to glob sql files: %v", err)
	}
	if len(files) != 24 {
		t.Errorf("expected 24 SQL migration files total (12 up + 12 down), found %d", len(files))
	}
}
