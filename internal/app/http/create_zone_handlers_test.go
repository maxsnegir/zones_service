package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storageMock "github.com/maxsnegir/zones_service/internal/app/http/mocks"
	"github.com/maxsnegir/zones_service/internal/domain/geojson"
)

func TestCreateZoneHandlerErr(t *testing.T) {
	type expectedResponse struct {
		ZoneId int    `json:"id"`
		Error  string `json:"error"`
	}
	tests := []struct {
		name               string
		data               []byte
		expectedStatusCode int
		expectedData       expectedResponse
	}{
		{
			name: "empty body",
			data: []byte{},
			expectedData: expectedResponse{
				Error: geojson.SerializationErr.Error(),
			},
		},
		{
			name: "wrong feature collection type",
			data: []byte(`{"type": "NotFeatureCollection", "features": []}`),
			expectedData: expectedResponse{
				Error: geojson.NotValidFeatureCollectionType{T: "NotFeatureCollection"}.Error(),
			},
		},
		{
			name: "features not passed",
			data: []byte(`{"type": "FeatureCollection"}`),
			expectedData: expectedResponse{
				Error: geojson.FeaturesIsRequiredErr.Error(),
			},
		},
		{
			name: "empty feature",
			data: []byte(`{"type": "FeatureCollection", "features": []}`),
			expectedData: expectedResponse{
				Error: geojson.FeaturesIsRequiredErr.Error(),
			},
		},
		{
			name: "empty geometry type",
			data: []byte(`{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {}}]}`),
			expectedData: expectedResponse{
				Error: geojson.GeometryTypeIsRequiredErr.Error(),
			},
		},
		{
			name: "empty coordinates",
			data: []byte(`{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "Polygon"}}]}`),
			expectedData: expectedResponse{
				Error: geojson.CoordinatesIsRequiredErr.Error(),
			},
		},
		{
			name: "wrong geometry type",
			data: []byte(`{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "NotPolygon", "coordinates": []}}]}`),
			expectedData: expectedResponse{
				Error: geojson.UnsupportedGeometryTypeErr{T: "NotPolygon"}.Error(),
			},
		},
		{
			name: "wrong polygon coordinates",
			data: []byte(`{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "Polygon", "coordinates": []}}]}`),
			expectedData: expectedResponse{
				Error: geojson.NotValidPolygonCoordinatesErr.Error(),
			},
		},
		{
			name: "wrong multipolygon coordinates",
			data: []byte(`{"type": "FeatureCollection", "features": [{"type": "Feature", "geometry": {"type": "MultiPolygon", "coordinates": [[]]}}]}`),
			expectedData: expectedResponse{
				Error: geojson.NotValidMultiPolygonCoordinatesErr.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			saverMock := storageMock.NewMockZoneSaver(ctrl)
			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, createZoneRoute, bytes.NewBuffer(tt.data))

			CreateZone(log, saverMock)(wr, req)
			response := wr.Result()
			defer response.Body.Close()

			var data expectedResponse
			if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
				t.Fatal(err)
			}

			require.Equal(t, http.StatusBadRequest, response.StatusCode)
			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
			require.Equal(t, data, tt.expectedData)
		})
	}
}

func TestCreateZoneHandler_Ok(t *testing.T) {
	type expectedResponse struct {
		ZoneId int    `json:"id"`
		Error  string `json:"error,omitempty"`
	}

	polygonGeoJson := `
	{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature", 
				"properties": {
					"color": "#ff0000"
				},
				"geometry": {
					"type": "Polygon", 
					"coordinates": [[[0, 0], [0, 1], [1, 1], [1, 0], [0, 0]]]
				}
			},
			{
				"type": "Feature",
				"properties": {
					"color": "#00ff00",
					"title": "Second Polygon"
				},
				"geometry": {
					"type": "Polygon",
					"coordinates": [[[2, 2], [2, 3], [3, 3], [3, 2], [2, 2]]]
				}
			}
		]
	}`
	multiPolygonGeoJson := `
	{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"properties": {
					"color": "#ff0000"
				},
				"geometry": {
					"type": "MultiPolygon",
					"coordinates": [[[[0, 0], [0, 1], [1, 1], [1, 0], [0, 0]]]]
				}
			},
			{
				"type": "Feature",
				"properties": {
					"color": "#00ff00"
				},
				"geometry": {
					"type": "MultiPolygon",
					"coordinates": [[[[2, 2], [2, 3], [3, 3], [3, 2], [2, 2]]]]
				}
			}
		]
	}`
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

	defer storage.CleanDB(context.Background())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, createZoneRoute, bytes.NewBuffer([]byte(tt.geoJson)))

			CreateZone(log, storage)(wr, req)
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

			zones, err := storage.GetZonesByIds(context.Background(), []int{tt.expectedId})
			assert.NoError(t, err)

			require.Equal(t, len(zones), 1)
			require.Equal(t, zones[0].ZoneId, tt.expectedId)

			var geoJsonFromDb geojson.FeatureCollectionJSON
			if err = json.Unmarshal([]byte(tt.geoJson), &geoJsonFromDb); err != nil {
				t.Fatal(err)
			}
			require.EqualValues(t, geoJsonFromDb, zones[0].GeoJSON)
		})
	}
}
