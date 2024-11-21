package main

import (
	"context"
	"github.com/vlks-dev/EffectiveMobileGoTest/config"
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
		return
	}
}
