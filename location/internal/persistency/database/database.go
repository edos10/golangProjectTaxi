package database

import (
	"context"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Init(ctx context.Context, config *DatabaseConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		logger.Error("Error with parsing DB config", zap.Error(err))
		return nil, err
	}

	pool, err := pgxpool.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		logger.Error("Error with connecting to DB", zap.Error(err))
		return nil, err
	}

	err = doMigration(config, logger)
	if err != nil {
		return nil, err
	}
	
	return pool, nil
}

func doMigration(config *DatabaseConfig, logger *zap.Logger) error {
	m, err := migrate.New(config.MigrationsDir, config.DSN)
	if err != nil {
		logger.Error("Error with creating DB migartion", zap.Error(err))
		return err
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Error with doing DB down migrations", zap.Error(err))
		return err
	}

	if err := m.Up(); err != nil {
		logger.Error("Error with doing DB up migrations", zap.Error(err))
		return err
	}

	return nil
}

