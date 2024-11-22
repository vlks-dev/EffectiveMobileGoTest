package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vlks-dev/EffectiveMobileGoTest/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunMigrations(ctx context.Context, logger *slog.Logger, config *config.Config) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	migrationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	connURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName,
	)

	db, err := sql.Open("pgx", connURL)
	if err != nil {
		logger.Error("failed to get postgres instance", "error", err.Error())
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("failed to close database connection", "error", closeErr.Error())
		}
	}()

	if err := db.PingContext(migrationCtx); err != nil {
		logger.Error("failed to ping postgres instance", "error", err.Error())
		return err
	}

	instance, err := pgx.WithInstance(db, &pgx.Config{
		MigrationsTable: "pgx_migrations",
		DatabaseName:    config.DBName,
	})
	if err != nil {
		logger.Error("failed to create pgx instance", "error", err.Error())
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/pgx_migrations/",
		config.DBName,
		instance,
	)
	if err != nil {
		logger.Error("failed to create migrate instance", "error", err.Error())
		return err
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil || dbErr != nil {
			logger.Error("failed to close migrate instance", "sourceErr", sourceErr.Error(), "dbErr", dbErr.Error())
		}
	}()

	errCh := make(chan error, 1)
	version, isDirty, _ := m.Version()

	go func() {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				logger.Debug("no migrations to apply")
				errCh <- nil
				return
			}

			logger.Error("failed to apply migrations", "error", err.Error(), "version", version, "isDirty", isDirty)

			if isDirty {
				logger.Debug("forcing migration cleanup")
				if forceErr := m.Force(int(version)); forceErr != nil {
					logger.Error("failed to force migration", "error", forceErr.Error())
					errCh <- forceErr
					return
				}
				errCh <- fmt.Errorf("migration version %d is dirty, error is %v", version, err.Error())
				return
			}

			// Попытка отката миграций
			logger.Warn("attempting rollback due to migration error")
			if rollbackErr := m.Steps(-1); rollbackErr != nil {
				logger.Error("failed to rollback migration", "error", rollbackErr.Error())
				errCh <- rollbackErr // Возвращаем ошибку отката, если она произошла
				return
			}

			logger.Info("rollback applied successfully")
			errCh <- err // Возвращаем исходную ошибку после успешного отката
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Warn("migration interrupted by shutdown signal")
		sourceErr, dbErr := m.Close()
		if sourceErr != nil || dbErr != nil {
			logger.Error("failed to close migrate instance", "sourceErr", sourceErr, "dbErr", dbErr)
		}
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return err
		}
		logger.Info(
			"migrations applied successfully",
			"version", version,
		)
		return nil
	}
}
