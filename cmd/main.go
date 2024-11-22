package main

import (
	"context"
	"github.com/vlks-dev/EffectiveMobileGoTest/config"
	"github.com/vlks-dev/EffectiveMobileGoTest/database"
	"github.com/vlks-dev/EffectiveMobileGoTest/migrations"
	"github.com/vlks-dev/EffectiveMobileGoTest/shared/logger"
)

func main() {
	cfg := config.LoadConfig()
	slog := logger.NewSlog()
	slog.Debug("config and logger initialized",
		"configuration", cfg)
	ctx := context.Background()
	err := migrations.RunMigrations(ctx, slog, cfg)
	if err != nil {
		slog.Error("migration failed", "error", err.Error())
		return
	}
	_, err = database.NewPostgresPool(cfg, slog, ctx)
	if err != nil {
		slog.Error("postgres pool failed", "error", err.Error())
		return
	}

}
