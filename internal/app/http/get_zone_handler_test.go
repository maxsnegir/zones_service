package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/dto"
	storageMock "github.com/maxsnegir/zones_service/internal/repository/mocks"
	"github.com/maxsnegir/zones_service/internal/service/zone"
)

func TestGetZonesByIds(t *testing.T) {
	ctx := context.Background()

	polygonId, err := createZoneFixture(ctx, polygonGeoJson)
	require.NoError(t, err)

	multiPolygonId, err := createZoneFixture(ctx, multiPolygonGeoJson)
	require.NoError(t, err)

	zonesCnt, err := storage.GetZonesCount(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, zonesCnt)

	type expectedResponse struct {
		ZoneId  int                        `json:"id"`
		GeoJSON *dto.FeatureCollectionJSON `json:"geojson"`
	}

	defer storage.CleanDB(ctx)

	tests := []struct {
		name             string
		zoneIds          []int
		expectedResponse []expectedResponse
	}{
		{
			name:    "get polygon",
			zoneIds: []int{polygonId},
			expectedResponse: []expectedResponse{
				{
					ZoneId:  polygonId,
					GeoJSON: dto.MustNewFeatureCollectionJSON(io.NopCloser(bytes.NewBufferString(polygonGeoJson))),
				},
			},
		},
		{
			name:    "get multipolygon",
			zoneIds: []int{multiPolygonId},
			expectedResponse: []expectedResponse{
				{
					ZoneId:  multiPolygonId,
					GeoJSON: dto.MustNewFeatureCollectionJSON(io.NopCloser(bytes.NewBufferString(multiPolygonGeoJson))),
				},
			},
		},
		{
			name:    "get all",
			zoneIds: []int{polygonId, multiPolygonId},
			expectedResponse: []expectedResponse{
				{
					ZoneId:  polygonId,
					GeoJSON: dto.MustNewFeatureCollectionJSON(io.NopCloser(bytes.NewBufferString(polygonGeoJson))),
				},
				{
					ZoneId:  multiPolygonId,
					GeoJSON: dto.MustNewFeatureCollectionJSON(io.NopCloser(bytes.NewBufferString(multiPolygonGeoJson))),
				},
			},
		},
		{
			name:             "get not existing",
			zoneIds:          []int{100, 200, 300, 400, 500},
			expectedResponse: []expectedResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zoneService := zone.New(log, storage, storage)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, getZonesRoute, nil)
			q := req.URL.Query()

			zoneIds := make([]string, len(tt.zoneIds))
			for i, zoneId := range tt.zoneIds {
				zoneIds[i] = strconv.Itoa(zoneId)
			}

			q.Add("ids", strings.Join(zoneIds, ","))
			req.URL.RawQuery = q.Encode()

			r.GetZones()(wr, req)
			response := wr.Result()
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
			require.Equal(t, http.StatusOK, response.StatusCode)

			var actualResponse []expectedResponse
			err = json.NewDecoder(response.Body).Decode(&actualResponse)
			require.NoError(t, err)

			require.Equal(t, len(tt.expectedResponse), len(actualResponse))

			for idx, resp := range actualResponse {
				require.Equal(t, tt.expectedResponse[idx].ZoneId, resp.ZoneId)

				var actualFeatureCollection geojson.FeatureCollection
				var expectedFeatureCollection geojson.FeatureCollection

				err = actualFeatureCollection.FromFeatureCollectionJSON(*resp.GeoJSON)
				require.NoError(t, err)
				err = expectedFeatureCollection.FromFeatureCollectionJSON(*tt.expectedResponse[idx].GeoJSON)
				require.NoError(t, err)

				require.EqualValues(t, expectedFeatureCollection, actualFeatureCollection)
			}
		})
	}
}

func TestGetZonesHandlerErr(t *testing.T) {
	type expectedResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name             string
		expectedResponse expectedResponse
		zoneIds          string
	}{
		{
			name: "empty ids",
			expectedResponse: expectedResponse{
				Error: ErrEmptyZoneIds.Error(),
			},
		},
		{
			name:    "wrong ids",
			zoneIds: "1,2,a,x,4",
			expectedResponse: expectedResponse{
				Error: ErrInvalidZoneId.Error(),
			},
		},
		{
			name:    "negative ids",
			zoneIds: "-1,2,3,4",
			expectedResponse: expectedResponse{
				Error: ErrInvalidZoneId.Error(),
			},
		},
		{
			name:    "zero ids",
			zoneIds: "0,2,3,4",
			expectedResponse: expectedResponse{
				Error: ErrInvalidZoneId.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zoneService := zone.New(log, storage, storage)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, getZonesRoute, nil)

			if tt.zoneIds != "" {
				q := req.URL.Query()
				q.Add("ids", tt.zoneIds)
				req.URL.RawQuery = q.Encode()
			}

			r.GetZones()(wr, req)
			response := wr.Result()
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
			require.Equal(t, http.StatusBadRequest, response.StatusCode)

			var actualResponse expectedResponse
			err := json.NewDecoder(response.Body).Decode(&actualResponse)
			require.NoError(t, err)
			require.EqualValues(t, tt.expectedResponse, actualResponse)

			zonesCnt, err := storage.GetZonesCount(context.Background())
			require.NoError(t, err)
			require.Equal(t, 0, zonesCnt)
		})
	}
}

func TestGetZonesHandler_DbErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSaver := storageMock.NewMockSaver(ctrl)
	mockProvider := storageMock.NewMockProvider(ctrl)

	zoneService := zone.New(log, mockSaver, mockProvider)
	r := NewRouter(mux.NewRouter(), zoneService, log)
	mockProvider.EXPECT().GetZonesByIds(gomock.Any(), gomock.Any()).Return(nil, errors.New("DB DOWN")).Times(1)

	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, getZonesRoute, nil)
	q := req.URL.Query()
	q.Add("ids", "1,2,3,4")
	req.URL.RawQuery = q.Encode()

	r.GetZones()(wr, req)
	response := wr.Result()
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, response.Header.Get("Content-Type"), "application/json")
	require.Equal(t, http.StatusInternalServerError, response.StatusCode)
}
