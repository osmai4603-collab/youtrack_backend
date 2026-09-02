package sqlstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RunMigrations ينفذ مخططات قاعدة البيانات (Migrations) للوصول إلى أحدث إصدار.
func (s *SqlStore) RunMigrations() error {
	_, sourceFile, _, _ := runtime.Caller(0)
	return s.RunMigrationsFrom(filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", "..", "..", "db", "migrations")))
}

// RunMigrationsFrom applies pending up migrations from a directory in filename order.
func (s *SqlStore) RunMigrationsFrom(directory string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	if _, err := s.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			version BIGINT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		version, err := migrationVersion(entry.Name())
		if err != nil {
			return err
		}

		var applied bool
		if err := s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, entry.Name()).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %d: %w", version, err)
		}
		if applied {
			continue
		}
		if version == 1 {
			var initialized bool
			if err := s.db.QueryRow(ctx, `SELECT to_regclass('public.users') IS NOT NULL`).Scan(&initialized); err != nil {
				return fmt.Errorf("check existing schema: %w", err)
			}
			if initialized {
				if _, err := s.db.Exec(ctx, `INSERT INTO schema_migrations (name, version) VALUES ($1, $2)`, entry.Name(), version); err != nil {
					return fmt.Errorf("baseline migration %d: %w", version, err)
				}
				continue
			}
		}

		sql, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", version, err)
		}
		if _, err = tx.Exec(ctx, string(sql)); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations (name, version) VALUES ($1, $2)`, entry.Name(), version)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
	}
	return nil
}

func migrationVersion(filename string) (int64, error) {
	prefix := strings.SplitN(filename, "_", 2)[0]
	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid migration filename %q: %w", filename, err)
	}
	return version, nil
}
