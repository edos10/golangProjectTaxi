package httpserver

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"moul.io/chizap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/location/internal/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server interface {
	Serve() error
	Shutdown(ctx context.Context)
}

type httpServer struct {
	config *HttpConfig
	locationService *service.LocationService

	server *http.Server

	writer *HttpWriter
	logger *zap.Logger
}

func (a *httpServer) Serve() error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(chizap.New(a.logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))
	
	si := NewServerInterface(a.locationService, a.writer)
	HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
		BaseURL: a.config.BasePath,
	})

	a.server = &http.Server{Addr: a.config.ServeAddress, Handler: r}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9000", nil)

	return a.server.ListenAndServe()
}

func (a *httpServer) Shutdown(ctx context.Context) {
	_ = a.server.Shutdown(ctx)
}

func New(
	config *HttpConfig,
	locationService *service.LocationService,
	logger *zap.Logger) Server {

	return &httpServer{
		config:      config,
		locationService: locationService,
		writer:      NewHttpWriter(logger),
		logger:      logger,
	}
}