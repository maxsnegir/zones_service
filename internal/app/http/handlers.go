package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/repository/psql"
)

type ZoneSaver interface {
	SaveZoneFromFeatureCollection(ctx context.Context, featureCollection geojson.FeatureCollection) (int, error)
}

type ZoneProvider interface {
	GetZonesByIds(ctx context.Context, ids []int) ([]*geojson.ZoneGEOJSON, error)
}

func CreateZone(log *slog.Logger, zoneSaver ZoneSaver) http.HandlerFunc {
	const op = "handlers.CreateZone"

	type ResponseData struct {
		ZoneId int    `json:"id,omitempty"`
		Error  string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var featureCollectionJSON geojson.FeatureCollectionJSON
		var responseData ResponseData

		if err := json.NewDecoder(r.Body).Decode(&featureCollectionJSON); err != nil {
			responseData.Error = geojson.SerializationErr.Error()
			writeResponse(w, http.StatusBadRequest, responseData)
			return
		}

		var featureCollection geojson.FeatureCollection
		if err := featureCollection.FromFeatureCollectionJSON(featureCollectionJSON); err != nil {
			responseData.Error = err.Error()
			writeResponse(w, http.StatusBadRequest, responseData)
			return
		}

		zoneId, err := zoneSaver.SaveZoneFromFeatureCollection(r.Context(), featureCollection)
		if err != nil {
			var e psql.PostgisValidationErr
			if errors.As(err, &e) {
				responseData.Error = e.Message
				writeResponse(w, http.StatusBadRequest, responseData)
				return
			}

			log.Error(fmt.Sprintf("%s: %v", op, err))
			writeResponse(w, http.StatusInternalServerError, nil)
			return
		}
		responseData.ZoneId = zoneId
		writeResponse(w, http.StatusCreated, responseData)
	}
}

func GetZones(log *slog.Logger, zoneProvider ZoneProvider) http.HandlerFunc {
	const op = "handlers.GetZones"

	type ResponseData struct {
		Error string `json:"error,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var responseData ResponseData

		w.Header().Set("Content-Type", "application/json")

		zoneIds, err := parseZoneIds(r.URL.Query().Get("ids"), true)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			responseData.Error = err.Error()
			if err := json.NewEncoder(w).Encode(responseData); err != nil {
				log.Error("%s: error on encode response: %v", op, err)
			}
		}

		zones, err := zoneProvider.GetZonesByIds(r.Context(), zoneIds)
		if err != nil {
			log.Error(fmt.Sprintf("%s: %v", op, err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(zones); err != nil {
			log.Error("%s: error on encode response: %v", op, err)
		}
	}
}
