package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/location/internal/logger"
	"github.com/location/internal/app"
)

func getConfigPath() string {
	var configPath string

	flag.StringVar(&configPath, "c", "../../config/settings.env.dev", "path to config file")
	flag.Parse()

	return configPath
}

func main() {
	config, err := app.NewConfig(getConfigPath())
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logger.GetLogger(false)
	if err != nil {
		log.Fatal(err)
	}

	app, err := app.New(config, logger)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		v := recover()

		if v != nil {
			ctx, _ := context.WithTimeout(ctx, 3*time.Second)
			app.Stop(ctx)
			log.Fatal(v)
		}
	}()

	app.Serve()

	<-ctx.Done()

	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	app.Stop(ctx)
}
