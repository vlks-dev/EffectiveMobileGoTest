package dbutil

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/vlks-dev/EffectiveMobileGoTest/utils/config"
	"log/slog"
)

type MigrationUtility struct {
	config *config.Config
	logger *slog.Logger
}

func NewMigrationUtility(config *config.Config, logger *slog.Logger) *MigrationUtility {
	return &MigrationUtility{config, logger}
}

func (m *MigrationUtility) GetInstance() (*sql.DB, error) {
	connURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		m.config.DBUser, m.config.DBPassword, m.config.DBHost, m.config.DBPort, m.config.DBName,
	)

	instance, dbErr := sql.Open("pgx", connURL)
	if dbErr != nil {
		m.logger.Error("failed to get postgres dbDrive",
			"connectionURL", connURL,
			"error", dbErr.Error(),
		)
		return nil, dbErr
	}

	return instance, nil
}

func (m *MigrationUtility) WithInstance(instance *sql.DB) (*migrate.Migrate, error) {
	dbDrive, dbError := pgx.WithInstance(instance, &pgx.Config{
		MigrationsTable: "pgx_migrations",
		DatabaseName:    m.config.DBName,
	})
	if dbError != nil {
		m.logger.Error("failed to create db instance", "error", dbError.Error())
		return nil, dbError
	}

	migration, mInstanceErr := migrate.NewWithDatabaseInstance(
		"file://migrations/pgx_migrations/",
		m.config.DBName,
		dbDrive,
	)
	if mInstanceErr != nil {
		m.logger.Error("failed to create migrate instance", "error", mInstanceErr.Error())
		return nil, mInstanceErr
	}
	return migration, nil
}
