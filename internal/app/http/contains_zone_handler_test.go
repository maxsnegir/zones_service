package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	storageMocks "github.com/maxsnegir/zones_service/internal/repository/mocks"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/dto"
	"github.com/maxsnegir/zones_service/internal/service/zone"
)

func TestContainsPoint_Ok(t *testing.T) {
	ctx := context.Background()

	polygonZoneId, err := createZoneFixture(ctx, polygonGeoJson)
	require.NoError(t, err)
	multiPolygonZoneId, err := createZoneFixture(ctx, multiPolygonGeoJson)
	require.NoError(t, err)

	tests := []struct {
		name     string
		request  dto.ZoneContainsPointIn
		expected []dto.ZoneContainsPointOut
	}{
		{
			name: "polygon: first feature contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{polygonZoneId},
				Point:   dto.Point{Lon: 0.6336, Lat: 0.5439},
			},
			expected: []dto.ZoneContainsPointOut{{ZoneId: polygonZoneId, Contains: true}},
		},
		{
			name: "polygon: second feature contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{polygonZoneId},
				Point:   dto.Point{Lon: 2.5448, Lat: 2.6211},
			},
			expected: []dto.ZoneContainsPointOut{{ZoneId: polygonZoneId, Contains: true}},
		},
		{
			name: "polygon: dont contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{polygonZoneId},
				Point:   dto.Point{Lon: 2.4728, Lat: 1.6995},
			},
			expected: []dto.ZoneContainsPointOut{{ZoneId: polygonZoneId, Contains: false}},
		},
		{
			name: "multipolygon: first polygon contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{multiPolygonZoneId},
				Point:   dto.Point{Lon: 0.6336, Lat: 0.5439},
			},
			expected: []dto.ZoneContainsPointOut{{ZoneId: multiPolygonZoneId, Contains: true}},
		},
		{
			name: "multipolygon: first polygon contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{multiPolygonZoneId},
				Point:   dto.Point{Lon: 2.5448, Lat: 2.6211},
			},
			expected: []dto.ZoneContainsPointOut{{ZoneId: multiPolygonZoneId, Contains: true}},
		},
		{
			name: "multipolygon: dont contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{multiPolygonZoneId},
				Point:   dto.Point{Lon: 2.4728, Lat: 1.6995},
			},
			expected: []dto.ZoneContainsPointOut{{ZoneId: multiPolygonZoneId, Contains: false}},
		},
		{
			name: "polygon and multipolygon contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{polygonZoneId, multiPolygonZoneId},
				Point:   dto.Point{Lon: 0.6336, Lat: 0.5439},
			},
			expected: []dto.ZoneContainsPointOut{
				{ZoneId: polygonZoneId, Contains: true},
				{ZoneId: multiPolygonZoneId, Contains: true},
			},
		},
		{
			name: "polygon and multipolygon dont contains point",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{polygonZoneId, multiPolygonZoneId},
				Point:   dto.Point{Lon: 2.4728, Lat: 1.6995},
			},
			expected: []dto.ZoneContainsPointOut{
				{ZoneId: polygonZoneId, Contains: false},
				{ZoneId: multiPolygonZoneId, Contains: false},
			},
		},
		{
			name: "not existing ids",
			request: dto.ZoneContainsPointIn{
				ZoneIds: []int{100, 1001},
				Point:   dto.Point{Lon: 2.4728, Lat: 1.6995},
			},
			expected: []dto.ZoneContainsPointOut{},
		},
	}

	defer storage.CleanDB(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zoneService := zone.New(log, storage, storage, storage)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			rawRequest, err := json.Marshal(tt.request)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, zonesContainsPoint, bytes.NewBuffer(rawRequest))

			r.ZonesContainsPoint()(w, req)

			response := w.Result()
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
			require.Equal(t, http.StatusOK, response.StatusCode)

			var actual []dto.ZoneContainsPointOut
			err = json.NewDecoder(response.Body).Decode(&actual)
			require.NoError(t, err)
			require.Equal(t, len(tt.expected), len(actual))
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestContainsPoint_Err(t *testing.T) {

	type errResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name               string
		requestData        string
		dbErr              bool
		expectedResponse   errResponse
		expectedStatusCode int
	}{
		{
			name:               "wrong body",
			requestData:        `{"ids": ["a", 1]}`,
			expectedResponse:   errResponse{Error: geojson.SerializationErr.Error()},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "empty ids",
			requestData:        `{"ids": [], "point": {"lon": 0, "lat": 0}}`,
			expectedResponse:   errResponse{Error: dto.EmptyIdsErr.Error()},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "wrong ids",
			requestData:        `{"ids": [-1, 1], "point": {"lon": 0, "lat": 0}}`,
			expectedResponse:   errResponse{Error: dto.ErrInvalidId.Error()},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "wrong point lat",
			requestData:        `{"ids": [1, 2], "point": {"lon": 0, "lat": 91}}`,
			expectedResponse:   errResponse{Error: dto.InvalidLatitudeError.Error()},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "wrong point lon",
			requestData:        `{"ids": [1, 2], "point": {"lon": -181, "lat": 90}}`,
			expectedResponse:   errResponse{Error: dto.InvalidLongitudeError.Error()},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "db error",
			requestData:        `{"ids": [1, 2], "point": {"lon": 0, "lat": 0}}`,
			dbErr:              true,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSaver := storageMocks.NewMockSaver(ctrl)
			mockProvider := storageMocks.NewMockProvider(ctrl)
			mockDeleter := storageMocks.NewMockDeleter(ctrl)

			if tt.dbErr == true {
				mockProvider.EXPECT().
					ContainsPoint(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]dto.ZoneContainsPointOut{}, errors.New("DB DOWN")).
					Times(1)
			}

			zoneService := zone.New(log, mockSaver, mockProvider, mockDeleter)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, zonesContainsPoint, bytes.NewBuffer([]byte(tt.requestData)))

			r.ZonesContainsPoint()(w, req)

			response := w.Result()
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
			require.Equal(t, tt.expectedStatusCode, response.StatusCode)

			if !tt.dbErr {
				var actual errResponse
				err := json.NewDecoder(response.Body).Decode(&actual)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResponse, actual)
			}
		})
	}
}
