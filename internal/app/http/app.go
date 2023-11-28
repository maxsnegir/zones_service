package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	log    *slog.Logger
	router *Router
	host   string
	port   int
}

func New(router *Router, host string, port int, log *slog.Logger) *App {
	return &App{
		router: router,
		host:   host,
		port:   port,
		log:    log,
	}
}

func (a *App) MustRun() {
	const op = "http.MustRun"

	if err := a.Run(); err != nil {
		panic(fmt.Sprintf("%s: %s", op, err))
	}
}

func (a *App) Run() error {
	const op = "http.Run"

	srv := &http.Server{
		Handler: a.router,
		Addr:    fmt.Sprintf("%s:%d", a.host, a.port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log := a.log.With(slog.String("op", op))
	log.Info("Starting HTTP server", slog.String("addr", fmt.Sprintf("http://%s:%d", a.host, a.port)))
	err := srv.ListenAndServe()
	return fmt.Errorf("%s: %w", op, err)
}

func (a *App) Stop() {
	const op = "http.Stop"

	a.log.With(slog.String("op", op)).Info("Gracefully stopped")
}
