package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vlks-dev/EffectiveMobileGoTest/config"
	"log/slog"
	"net/url"
)

type PostgresPool struct {
	pool *pgxpool.Pool
}

func NewPostgresPool(config *config.Config, logger *slog.Logger, ctx context.Context) (*pgxpool.Pool, error) {
	connURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName,
	)

	parseUrl, err := url.Parse(connURL)
	if err != nil {
		logger.Error("failed to parse postgres connection url", "error", err.Error())
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		logger.Error("unable to parse pool config", "error", err.Error(),
			"Host", parseUrl.Host,
			"dbName", parseUrl.Path,
			"err", err.Error(),
		)
		return nil, err
	}
	poolConfig.MaxConnIdleTime = config.DBMaxIdle
	poolConfig.MaxConnLifetime = config.DBMaxConn

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("unable to connect to postgres pool", "error", err.Error())
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping postgres pool", "error", err.Error())
		return nil, err
	}
	logger.Info("successfully connected to postgres pool",
		"Host", parseUrl.Host,
		"dbName", parseUrl.Path,
		"max idle time", poolConfig.MaxConnIdleTime,
		"max conn lifetime", poolConfig.MaxConnLifetime,
	)

	return pool, nil
}
