// Package app is the application bootstrapping and composition root.
package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandlers "youtrack_backend/internal/httphandlers"
	"youtrack_backend/internal/httphandlers/middlewares"
	"youtrack_backend/internal/infrastructure/health"
	"youtrack_backend/internal/infrastructure/postgres"
	"youtrack_backend/internal/infrastructure/server"
	"youtrack_backend/internal/infrastructure/worker"
	"youtrack_backend/internal/platform/config"
	platformctx "youtrack_backend/internal/platform/context"
	platformlogger "youtrack_backend/internal/platform/logger"
	"youtrack_backend/internal/services/auth"
	perms2 "youtrack_backend/internal/services/permissions_services"
	"youtrack_backend/internal/services/rbac"
	"youtrack_backend/internal/storage"
	authstorage "youtrack_backend/internal/storage/auth"
	permstorage "youtrack_backend/internal/storage/permissions"
	"youtrack_backend/internal/usecases"
	"youtrack_backend/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultConfigPath = "config/config.json"

// App is the composition root coordinating initialization and cleanup order.
type App struct {
	cfg       *config.Configuration
	logger    *slog.Logger
	logCloser io.Closer
	health    *health.HealthChecker
	workerMgr *worker.Manager
	srv       *server.Server
	pool      *pgxpool.Pool
	resource  io.Closer
}

// New returns an uninitialized App; call Bootstrap then Run.
func New() *App {
	return &App{}
}

// Bootstrap performs Phase 1 (Initialization) and Phase 3 (Configuration).
func (a *App) Bootstrap(args []string) (err error) {
	// PHASE 1a: Configuration
	engine := config.NewEngine(
		config.WithConfigFile(defaultConfigPath),
		config.WithCLIArgs(args),
	)
	cfg, err := engine.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	a.cfg = cfg

	// PHASE 1b: Multi-target structured logger
	log, logCloser := newLogger(cfg.Logger)
	a.logger = log
	a.logCloser = logCloser
	defer func() {
		if err != nil {
			_ = a.logCloser.Close()
		}
	}()

	a.logger.Info("configuration loaded",
		slog.String("interface", cfg.Server.Host),
		slog.String("port", cfg.Server.Port),
		slog.String("log_level", cfg.Logger.Level),
		slog.Bool("log_to_file", cfg.Logger.ToFile),
	)

	// PHASE 1c: Infrastructure (fail fast)
	switch cfg.Database.Driver {
	case "postgres":
		if err := a.connectPostgres(); err != nil {
			return err
		}
	case "memory":
		a.logger.Info("using in-memory development database driver")
		a.health = health.NewHealthChecker(noopPinger{})
	default:
		return fmt.Errorf("unsupported database driver %q", cfg.Database.Driver)
	}

	// PHASE 3: Configuration — workers, router, middleware chain, server
	a.workerMgr = worker.NewManager(a.logger)

	if a.pool != nil {
		a.workerMgr.Start("pool-monitor", newPoolMonitor(a.pool, a.logger))
	}

	repositories := storage.NewRepositoriesFromPostgres(a.pool)

	a.workerMgr.Start("session-purge", newSessionPurger(repositories.Sessions, a.logger))

	authenticator := auth.NewCookieSessionResolver(repositories.Sessions, repositories.Permissions)
	protect := middlewares.Authenticate(authenticator)

	// RBAC Engine initialization (ANSI/INCITS 359 Flat RBAC + ABAC + SoD)
	var rbacOptions []rbac.ServiceOption
	var roleAdapter rbac.RoleRepository
	if a.pool != nil {
		pgRoleAdapter := permstorage.NewPostgresRBACRoleAdapter(a.pool)
		sessionAdapter := authstorage.NewPostgresRBACSessionAdapter(a.pool)
		roleAdapter = pgRoleAdapter
		rbacOptions = append(rbacOptions,
			rbac.WithRoleRegistryRepository(pgRoleAdapter),
			rbac.WithSessionRepository(sessionAdapter),
		)
	}

	catalogSvc := perms2.NewService()
	catalogAdapter := rbac.NewPermissionCatalogAdapter(catalogSvc)

	rbacSvc := rbac.NewService(roleAdapter, catalogAdapter, catalogSvc, rbacOptions...)
	authorizer := rbac.NewAuthorizerAdapter(rbacSvc)

	sessionTTL := cfg.Auth.RefreshTTL.Duration()
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}

	usecases := usecases.New(repositories, sessionTTL, a.logger, repositories.Groups, repositories.Users, repositories.Permissions, rbacSvc)
	usecases.SetRBAC(rbacSvc)

	handlers := httphandlers.NewYouTrackHandlers(usecases, a.logger)

	router := httphandlers.NewYouTrackRouter(handlers, a.health, a.logger, protect, authorizer)

	handler := platformctx.Middleware(
		platformlogger.HTTPLoggingMiddleware(
			a.logger,
			platformlogger.WithProbeSuppression("/livez", "/readyz"),
		)(router.Base),
	)

	var resources []io.Closer
	if a.resource != nil {
		resources = append(resources, a.resource)
	}
	a.srv = server.New(cfg.Server, handler, a.health, a.workerMgr, a.logger, resources...)

	a.logger.Info("bootstrap complete",
		slog.String("addr", cfg.Server.Address()),
		slog.String("driver", cfg.Database.Driver),
	)
	return nil
}

const (
	connectTimeout   = 10 * time.Second
	migrationTimeout = 5 * time.Minute
)

// connectPostgres establishes the connection pool and applies schema migrations.
func (a *App) connectPostgres() error {
	connectCtx, cancelConnect := context.WithTimeout(context.Background(), connectTimeout)
	defer cancelConnect()

	pool, err := postgres.NewPool(connectCtx, postgres.NewConfigFromSettings(a.cfg.Database), a.logger)
	if err != nil {
		return fmt.Errorf("initialize postgres: %w", err)
	}
	a.pool = pool
	a.resource = poolCloser{pool: pool}

	migrateCtx, cancelMigrate := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancelMigrate()

	if err := postgres.RunMigrations(migrateCtx, pool, migrations.FS, ".", a.logger); err != nil {
		pool.Close()
		a.pool = nil
		a.resource = nil
		return fmt.Errorf("run postgres migrations: %w", err)
	}

	a.health = health.NewHealthChecker(postgres.NewHealthChecker(pool))
	return nil
}

// Run executes the serving lifecycle.
func (a *App) Run() error {
	if a.srv == nil {
		return fmt.Errorf("app has not been bootstrapped")
	}

	sigCtx, sigStop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer sigStop()

	a.logger.Info("starting service", slog.String("addr", a.cfg.Server.Address()))

	if err := a.srv.Run(sigCtx); err != nil {
		a.logger.Error("server runtime failure", slog.Any("error", err))
		return fmt.Errorf("server runtime failure: %w", err)
	}

	a.logger.Info("service terminated successfully")

	if a.logCloser != nil {
		if err := a.logCloser.Close(); err != nil {
			a.logger.Warn("failed to close log files", slog.Any("error", err))
		}
	}
	return nil
}

// Close flushes and closes application-level resources.
func (a *App) Close() error {
	if a.logCloser != nil {
		return a.logCloser.Close()
	}
	return nil
}

type noopPinger struct{}

func (noopPinger) Ping(context.Context) error { return nil }

type poolCloser struct {
	pool *pgxpool.Pool
}

func (c poolCloser) Close() error {
	c.pool.Close()
	return nil
}

// newSessionPurger periodically removes expired sessions from storage.
func newSessionPurger(sessions authstorage.SessionRepository, logger *slog.Logger) func(ctx context.Context) {
	interval := 1 * time.Hour

	return func(ctx context.Context) {
		if sessions == nil {
			return
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := sessions.PurgeExpired(ctx); err != nil {
					logger.LogAttrs(ctx, slog.LevelError, "session purge failed",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}
}

func newPoolMonitor(pool *pgxpool.Pool, logger *slog.Logger) func(ctx context.Context) {
	hc := postgres.NewHealthChecker(pool)
	interval := 30 * time.Second

	return func(ctx context.Context) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !logger.Enabled(ctx, slog.LevelDebug) {
					continue
				}
				stats := hc.Stats()
				logger.LogAttrs(ctx, slog.LevelDebug, "postgres pool telemetry",
					slog.Int("total_conns", int(stats.TotalConns)),
					slog.Int("idle_conns", int(stats.IdleConns)),
					slog.Int("acquired_conns", int(stats.AcquiredConns)),
					slog.Int64("acquire_count", stats.AcquireCount),
				)
			}
		}
	}
}
