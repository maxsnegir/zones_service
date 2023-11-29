package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	createZoneRoute = "/zones"
	getZonesRoute   = "/zones"
)

type Router struct {
	router *mux.Router
	log    *slog.Logger
}

func NewRouter(router *mux.Router, logger *slog.Logger) *Router {
	return &Router{
		router: router,
		log:    logger,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *Router) ConfigureRouter(zoneSaver ZoneSaver, zoneProvider ZoneProvider) {
	// Routes
	r.router.HandleFunc(createZoneRoute, CreateZone(r.log, zoneSaver)).Methods(http.MethodPost)
	r.router.HandleFunc(getZonesRoute, GetZones(r.log, zoneProvider)).Methods(http.MethodGet)

	// Middlewares
	r.router.Use(r.loggingMiddleware)
}

func (r *Router) loggingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		r.log.Info(fmt.Sprintf("%s %s", request.Method, request.URL.String()))

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, request)
	})
}
