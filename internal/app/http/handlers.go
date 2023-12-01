package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/repository/psql"
)

func (r *Router) CreateZone() http.HandlerFunc {
	const op = "handlers.CreateZone"

	type ResponseData struct {
		ZoneId int    `json:"id,omitempty"`
		Error  string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, req *http.Request) {
		var responseData ResponseData

		featureCollectionJSON, err := geojson.NewFeatureCollectionJSON(req.Body)
		if err != nil {
			responseData.Error = geojson.SerializationErr.Error()
			r.JsonResponse(w, http.StatusBadRequest, responseData)
			return
		}

		var featureCollection geojson.FeatureCollection
		if err := featureCollection.FromFeatureCollectionJSON(*featureCollectionJSON); err != nil {
			responseData.Error = err.Error()
			r.JsonResponse(w, http.StatusBadRequest, responseData)
			return
		}

		zoneId, err := r.ZoneService.SaveZoneFromFeatureCollection(req.Context(), featureCollection)
		if err != nil {
			var e psql.PostgisValidationErr
			if errors.As(err, &e) {
				responseData.Error = e.Message
				r.JsonResponse(w, http.StatusBadRequest, responseData)
				return
			}

			r.log.Error(fmt.Sprintf("%s: %v", op, err))
			r.JsonResponse(w, http.StatusInternalServerError, nil)
			return
		}
		responseData.ZoneId = zoneId
		r.JsonResponse(w, http.StatusCreated, responseData)
	}
}

func (r *Router) GetZones() http.HandlerFunc {
	const op = "handlers.GetZones"

	type ErrResponseData struct {
		Error string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, req *http.Request) {

		zoneIds, err := parseZoneIds(req.URL.Query().Get("ids"), true)
		if err != nil {
			responseData := ErrResponseData{Error: err.Error()}
			r.JsonResponse(w, http.StatusBadRequest, responseData)
			return
		}

		zones, err := r.ZoneService.GetZonesByIds(req.Context(), zoneIds)
		if err != nil {
			r.log.Error(fmt.Sprintf("%s: %v", op, err))
			r.JsonResponse(w, http.StatusInternalServerError, nil)
			return
		}

		r.JsonResponse(w, http.StatusOK, zones)
	}
}
