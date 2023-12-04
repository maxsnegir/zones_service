package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/dto"
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

		featureCollectionJSON, err := dto.NewFeatureCollectionJSON(req.Body)
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

func (r *Router) ZonesContainsPoint() http.HandlerFunc {
	const op = "handlers.ZonesContainsPoint"

	type ErrResponseData struct {
		Error string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, req *http.Request) {
		var requestData dto.ZoneContainsPointIn

		if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
			response := ErrResponseData{Error: geojson.SerializationErr.Error()}
			r.JsonResponse(w, http.StatusBadRequest, response)
			return
		}
		if err := requestData.Validate(); err != nil {
			response := ErrResponseData{Error: err.Error()}
			r.JsonResponse(w, http.StatusBadRequest, response)
			return
		}

		result, err := r.ZoneService.ContainsPoint(req.Context(), requestData)
		if err != nil {
			r.log.Error(fmt.Sprintf("%s: %v", op, err))
			r.JsonResponse(w, http.StatusInternalServerError, nil)
			return
		}

		r.JsonResponse(w, http.StatusOK, result)
	}
}

func (r *Router) DeleteZone() http.HandlerFunc {
	const op = "handlers.DeleteZone"

	type errResponseData struct {
		Error string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, req *http.Request) {
		idStr := mux.Vars(req)["id"]

		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			response := errResponseData{Error: "invalid zone id"}
			r.JsonResponse(w, http.StatusBadRequest, response)
			return
		}

		err = r.ZoneService.DeleteZone(req.Context(), id)
		if err != nil {
			r.log.Error(fmt.Sprintf("%s: %v", op, err))
			r.JsonResponse(w, http.StatusInternalServerError, nil)
			return
		}

		r.JsonResponse(w, http.StatusNoContent, nil)
	}
}
