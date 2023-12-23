package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"driver_service/internal/app"
	"driver_service/internal/config"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "/internal/config.yaml", "set config path")
	flag.Parse()

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		fmt.Println(fmt.Errorf("fatal: init config %w", err))
		os.Exit(1)
	}

	a := app.NewAppServer(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		v := recover()

		if v != nil {
			ctx, _ := context.WithTimeout(ctx, 3*time.Second)
			a.Stop(ctx)
			fmt.Println(v)
			os.Exit(1)
		}
	}()

	a.Run()

	<-ctx.Done()
	ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	a.Stop(ctx)

}
