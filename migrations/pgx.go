package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vlks-dev/EffectiveMobileGoTest/utils/config"
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

	db, dbErr := sql.Open("pgx", connURL)
	if dbErr != nil {
		logger.Error("failed to get postgres instance",
			"connectionURL", connURL,
			"error", dbErr.Error(),
		)
		return dbErr
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("failed to close database connection", "error", closeErr.Error())
		}
	}()

	if pingErr := db.PingContext(migrationCtx); pingErr != nil {
		logger.Error("failed to ping postgres instance", "error", pingErr.Error())
		return pingErr
	}

	instance, instanceErr := pgx.WithInstance(db, &pgx.Config{
		MigrationsTable: "pgx_migrations",
		DatabaseName:    config.DBName,
	})
	if instanceErr != nil {
		logger.Error("failed to create pgx instance", "error", instanceErr.Error())
		return instanceErr
	}

	m, mInstanceErr := migrate.NewWithDatabaseInstance(
		"file://migrations/pgx_migrations/",
		config.DBName,
		instance,
	)
	if mInstanceErr != nil {
		logger.Error("failed to create migrate instance", "error", mInstanceErr.Error())
		return mInstanceErr
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil || dbErr != nil {
			logger.Error("failed to close migrate instance", "sourceErr", sourceErr.Error(), "dbErr", dbErr.Error())
		}
	}()

	upErr := m.Up()

	version, isDirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		logger.Error("failed to get migrate version", "error", err.Error())
		return err
	}
	if isDirty {
		logger.Warn("Attempting rollback for dirty migration", "version", version)
		if rollbackErr := m.Steps(-1); rollbackErr != nil {
			logger.Error("Failed to rollback migration", "error", rollbackErr.Error())
			if forceRollbackErr := m.Force(int(version)); forceRollbackErr != nil {
				logger.Error("Failed to force rollback", "error", forceRollbackErr.Error())
				return forceRollbackErr
			}
			logger.Info("Force rollback successful", "version", version, "dirty", isDirty)
		}
	}
	if upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		if strings.Contains(upErr.Error(), "42P07") {
			re := regexp.MustCompile(`"([^"]+)"`)
			logger.Warn("Table already exists, forcing migration",
				"table", re.FindStringSubmatch(upErr.Error())[1],
				"version", version,
			)
			if err = m.Force(int(version)); err != nil {
				logger.Error("failed to force migration", "error", err.Error())
				return err
			}
			logger.Info("Force migration successful", "version", version)
			return nil
		} else {
			logger.Error("failed to run up migration", "error", upErr.Error())
			return upErr
		}
	}
	version, _, _ = m.Version()
	if errors.Is(upErr, migrate.ErrNoChange) {
		logger.Info("migration is up to date", "version", version, "dirty", isDirty)
		return nil
	}
	logger.Info("migration successful!", "version", version)
	return nil
}
