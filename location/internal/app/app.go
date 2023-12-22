package app

import (
	"net/http"
	"context"
	"go.uber.org/zap"

	"github.com/location/internal/persistency"
	"github.com/location/internal/persistency/database"
	"github.com/location/internal/httpserver"
	"github.com/location/internal/services"
)

type App interface {
	Serve()
	Stop(ctx context.Context)
}

type app struct {
	config *Config
	logger *zap.Logger
	httpServer httpserver.Server
}

func (a *app) Serve() {
	go func() {
		if err := a.httpServer.Serve(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	a.logger.Info("Server started")
}

func (a *app) Stop(ctx context.Context) {
	a.httpServer.Shutdown(ctx)
}

func New(config *Config, logger *zap.Logger) (App, error) {
	ctx := context.Background()

	pgxPool, err := database.Init(ctx, &config.Database, logger)
	if err != nil {
		return nil, err
	}
	repo, err := persistency.NewDriverRepo(pgxPool, logger)
	if err != nil {
		return nil, err
	}

	service := service.NewLocationService(repo)
	server := httpserver.New(&config.HTTP, service, logger)

	a := &app{
		config:      config,
		logger: logger,
		httpServer: server,
	}

	return a, nil
}