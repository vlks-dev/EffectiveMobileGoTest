package migrations

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Migration struct {
	ctx       context.Context
	logger    *slog.Logger
	db        *sql.DB
	migration *migrate.Migrate
}

func Migrator(ctx context.Context, logger *slog.Logger, db *sql.DB, migration *migrate.Migrate) *Migration {
	return &Migration{ctx, logger, db, migration}
}

func (m *Migration) Run() error {
	ctx, stop := signal.NotifyContext(m.ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	migrationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if pingErr := m.db.PingContext(migrationCtx); pingErr != nil {
		m.logger.Error("failed to ping postgres instance", "error", pingErr.Error())
		return pingErr
	}

	upErr := m.migration.Up()

	version, isDirty, err := m.migration.Version()

	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		m.logger.Error("failed to get migrate version", "error", err.Error())
		return err
	}
	if isDirty {
		m.logger.Warn("Attempting rollback for dirty migration", "version", version)
		if rollbackErr := m.migration.Down(); rollbackErr != nil {
			m.logger.Error("Failed to rollback migration", "error", rollbackErr.Error())
			if forceRollbackErr := m.migration.Force(int(version)); forceRollbackErr != nil {
				m.logger.Error("Failed to force rollback", "error", forceRollbackErr.Error())
				return forceRollbackErr
			}
			m.logger.Info("Force rollback successful", "version", version, "dirty", isDirty)
		}
	}
	if upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		if strings.Contains(upErr.Error(), "42P07") {
			re := regexp.MustCompile(`"([^"]+)"`)
			m.logger.Warn("Table already exists, forcing migration",
				"table", re.FindStringSubmatch(upErr.Error())[1],
				"version", version,
			)

			if err = m.migration.Force(int(version)); err != nil {
				m.logger.Error("failed to force migration", "error", err.Error())
				return err
			}

			m.logger.Info("Force migration successful", "version", version)
			return nil
		} else {
			m.logger.Error("failed to run up migration", "error", upErr.Error())
			return upErr
		}
	}

	if errors.Is(upErr, migrate.ErrNoChange) {
		m.logger.Info("migration is up to date", "version", version, "dirty", isDirty)
		return nil
	}
	m.logger.Info("migration successful!", "version", version)
	return nil
}
