package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/maxsnegir/zones_service/internal/config"
	"github.com/maxsnegir/zones_service/internal/logger"
	"github.com/maxsnegir/zones_service/internal/repository/psql"
	"github.com/maxsnegir/zones_service/internal/service/zone"

	httpserver "github.com/maxsnegir/zones_service/internal/app/http"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)
	log.Debug("config: ", slog.Any("config", cfg))

	storage, err := psql.New(ctx, log, cfg.Storage.DSN)
	if err != nil {
		panic(err)
	}

	zoneService := zone.New(log, storage, storage)
	appRouter := httpserver.NewRouter(mux.NewRouter(), zoneService, log)
	appRouter.ConfigureRouter()

	app := httpserver.New(appRouter, cfg.Server.Host, cfg.Server.Port, log)

	go app.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	app.Stop()
	log.Info("Gracefully stopped")
}
