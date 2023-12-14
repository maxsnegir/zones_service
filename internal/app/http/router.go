package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/maxsnegir/zones_service/internal/service/zone"
)

const (
	createZoneRoute            = "/create"
	getZonesRoute              = "/get"
	zonesContainsPoint         = "/contains"
	anyZonesContainsPoint      = "/any_contains"
	batchAnyZonesContainsPoint = "/batch_any_contains"
	deleteZoneRoute            = "/delete/{id}"
)

type Router struct {
	router      *mux.Router
	log         *logrus.Logger
	ZoneService *zone.Service
}

func NewRouter(router *mux.Router, zoneService *zone.Service, logger *logrus.Logger) *Router {
	return &Router{
		router:      router,
		ZoneService: zoneService,
		log:         logger,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *Router) ConfigureRouter() {
	// Routes
	r.router.HandleFunc(createZoneRoute, r.CreateZone()).Methods(http.MethodPost)
	r.router.HandleFunc(getZonesRoute, r.GetZones()).Methods(http.MethodGet)
	r.router.HandleFunc(zonesContainsPoint, r.ZonesContainsPoint()).Methods(http.MethodPost)
	r.router.HandleFunc(anyZonesContainsPoint, r.AnyOfZonesContainsPint()).Methods(http.MethodPost)
	r.router.HandleFunc(batchAnyZonesContainsPoint, r.BatchAnyOfZonesContainsPint()).Methods(http.MethodPost)
	r.router.HandleFunc(deleteZoneRoute, r.DeleteZone()).Methods(http.MethodDelete)

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

func (r *Router) JsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	const op = "http.JsonResponse"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			r.log.Error(fmt.Sprintf("%s: %v", op, err))
		}
	}
}
