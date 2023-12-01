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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storageMock "github.com/maxsnegir/zones_service/internal/app/http/mocks"
	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/repository/psql"
	"github.com/maxsnegir/zones_service/internal/service/zone"
)

type expectedResponse struct {
	ZoneId int    `json:"id"`
	Error  string `json:"error,omitempty"`
}

func TestCreateZoneHandlerErr(t *testing.T) {
	tests := []struct {
		name               string
		data               string
		expectedStatusCode int
		expectedData       expectedResponse
	}{
		{
			name: "empty body",
			data: "",
			expectedData: expectedResponse{
				Error: geojson.SerializationErr.Error(),
			},
		},
		{
			name: "wrong feature collection type",
			data: `{"type": "NotFeatureCollection", "features": []}`,
			expectedData: expectedResponse{
				Error: geojson.NotValidFeatureCollectionType{T: "NotFeatureCollection"}.Error(),
			},
		},
		{
			name: "features not passed",
			data: `{"type": "FeatureCollection"}`,
			expectedData: expectedResponse{
				Error: geojson.FeaturesIsRequiredErr.Error(),
			},
		},
		{
			name: "empty feature",
			data: `{"type": "FeatureCollection", "features": []}`,
			expectedData: expectedResponse{
				Error: geojson.FeaturesIsRequiredErr.Error(),
			},
		},
		{
			name: "empty geometry type",
			data: `{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {}}]}`,
			expectedData: expectedResponse{
				Error: geojson.GeometryTypeIsRequiredErr.Error(),
			},
		},
		{
			name: "empty coordinates",
			data: `{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "Polygon"}}]}`,
			expectedData: expectedResponse{
				Error: geojson.CoordinatesIsRequiredErr.Error(),
			},
		},
		{
			name: "wrong geometry type",
			data: `{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "NotPolygon", "coordinates": []}}]}`,
			expectedData: expectedResponse{
				Error: geojson.UnsupportedGeometryTypeErr{T: "NotPolygon"}.Error(),
			},
		},
		{
			name: "wrong polygon coordinates format",
			data: `{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "Polygon", "coordinates": []}}]}`,
			expectedData: expectedResponse{
				Error: geojson.NotValidPolygonCoordinatesErr.Error(),
			},
		},
		{
			name: "wrong multipolygon coordinates format",
			data: `{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "MultiPolygon", "coordinates": [[]]}}]}`,
			expectedData: expectedResponse{
				Error: geojson.NotValidMultiPolygonCoordinatesErr.Error(),
			},
		},
		{
			name: "wrong polygon coordinates",
			data: `{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "Polygon", "coordinates": [[[1, 2]]]}}]}`,
			expectedData: expectedResponse{
				Error: psql.PostgisValidationErr{Message: "Polygon must have at least four points in each ring"}.Error(),
			},
		},
	}

	ctx := context.Background()
	defer storage.CleanDB(ctx)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zoneService := zone.New(log, storage, storage)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, createZoneRoute, bytes.NewBuffer([]byte(tt.data)))
			r.CreateZone()(wr, req)
			response := wr.Result()
			defer response.Body.Close()

			var data expectedResponse
			if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
				t.Fatal(err)
			}

			require.Equal(t, http.StatusBadRequest, response.StatusCode)
			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
			require.Equal(t, data, tt.expectedData)

			count, err := storage.GetZonesCount(ctx)
			assert.NoError(t, err)
			require.Equal(t, 0, count)
		})
	}
}

func TestCreateZoneHandler_Ok(t *testing.T) {

	tests := []struct {
		name       string
		geoJson    string
		expectedId int
	}{
		{
			name:       "polygon",
			geoJson:    polygonGeoJson,
			expectedId: 1,
		},
		{
			name:       "multipolygon",
			geoJson:    multiPolygonGeoJson,
			expectedId: 2,
		},
	}
	ctx := context.Background()
	defer storage.CleanDB(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zoneService := zone.New(log, storage, storage)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, createZoneRoute, bytes.NewBuffer([]byte(tt.geoJson)))
			r.CreateZone()(wr, req)
			response := wr.Result()
			defer response.Body.Close()

			require.Equal(t, http.StatusCreated, response.StatusCode)
			require.Equal(t, response.Header.Get("Content-Type"), "application/json")

			var data expectedResponse
			if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, data.ZoneId, tt.expectedId)
			require.Equal(t, data.Error, "")

			zones, err := storage.GetZonesByIds(ctx, []int{tt.expectedId})
			assert.NoError(t, err)

			require.Equal(t, len(zones), 1)
			require.Equal(t, zones[0].ZoneId, tt.expectedId)

			var geoJsonFromDb geojson.FeatureCollectionJSON
			if err = json.Unmarshal([]byte(tt.geoJson), &geoJsonFromDb); err != nil {
				t.Fatal(err)
			}
			require.EqualValues(t, geoJsonFromDb, zones[0].GeoJSON)

			count, err := storage.GetZonesCount(ctx)
			assert.NoError(t, err)
			require.Equal(t, tt.expectedId, count)
		})
	}
}

func TestCreateZoneHandler_DbErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSaver := storageMock.NewMockZoneSaver(ctrl)
	mockProvider := storageMock.NewMockZoneProvider(ctrl)
	mockSaver.EXPECT().SaveZoneFromFeatureCollection(gomock.Any(), gomock.Any()).Return(0, errors.New("DB DOWN")).Times(1)

	zoneService := zone.New(log, mockSaver, mockProvider)
	r := NewRouter(mux.NewRouter(), zoneService, log)

	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, createZoneRoute, bytes.NewBuffer([]byte(polygonGeoJson)))
	r.CreateZone()(wr, req)
	response := wr.Result()
	defer response.Body.Close()

	require.Equal(t, response.Header.Get("Content-Type"), "application/json")
	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
}
