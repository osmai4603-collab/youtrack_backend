// Command migrate applies all pending embedded database migrations and exits.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"youtrack_backend/internal/infrastructure/postgres"
	"youtrack_backend/internal/platform/config"
	"youtrack_backend/migrations"
)

const (
	configPath     = "config/config.json"
	connectTimeout = 10 * time.Second
	migrateTimeout = 5 * time.Minute
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.NewEngine(
		config.WithConfigFile(configPath),
		config.WithCLIArgs(os.Args[1:]),
	).Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if cfg.Database.Driver != "postgres" {
		return fmt.Errorf("migrations require postgres, got %q", cfg.Database.Driver)
	}

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), connectTimeout)
	defer cancelConnect()

	pool, err := postgres.NewPool(connectCtx, postgres.NewConfigFromSettings(cfg.Database), slog.Default())
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	migrateCtx, cancelMigrate := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancelMigrate()

	if err := postgres.RunMigrations(migrateCtx, pool, migrations.FS, ".", slog.Default()); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
