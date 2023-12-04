package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	mock_zone "github.com/maxsnegir/zones_service/internal/repository/mocks"
	"github.com/maxsnegir/zones_service/internal/service/zone"
)

func TestDeleteZoneHandler_Ok(t *testing.T) {
	ctx := context.Background()

	polygonZoneId, err := createZoneFixture(ctx, polygonGeoJson)
	require.NoError(t, err)
	multiPolygonZoneId, err := createZoneFixture(ctx, multiPolygonGeoJson)
	require.NoError(t, err)

	cnt, err := storage.GetZonesCount(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, cnt)

	zoneService := zone.New(log, storage, storage, storage)
	r := NewRouter(mux.NewRouter(), zoneService, log)

	zoneIds := []int{polygonZoneId, multiPolygonZoneId}

	defer storage.CleanDB(ctx)

	for _, zoneId := range zoneIds {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/delete/", nil)

		req = mux.SetURLVars(req, map[string]string{"id": fmt.Sprintf("%d", zoneId)})
		r.DeleteZone()(w, req)
		response := w.Result()

		require.Equal(t, response.Header.Get("Content-Type"), "application/json")
		require.Equal(t, http.StatusNoContent, response.StatusCode)

		actualCnt, err := storage.GetZonesCount(ctx)
		require.NoError(t, err)
		require.Equal(t, cnt-1, actualCnt)
		cnt = actualCnt
	}
}

func TestDeleteZoneHandler_Err(t *testing.T) {
	tests := []struct {
		name               string
		id                 string
		expectedStatusCode int
		dbErr              bool
	}{
		{
			name:               "empty id",
			id:                 "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "wrong id",
			id:                 "a",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "negative id",
			id:                 "-1",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "zero id",
			id:                 "0",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "db error",
			id:                 "1",
			expectedStatusCode: http.StatusInternalServerError,
			dbErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			saver := mock_zone.NewMockSaver(ctrl)
			provider := mock_zone.NewMockProvider(ctrl)
			deleter := mock_zone.NewMockDeleter(ctrl)

			if tt.dbErr == true {
				deleter.EXPECT().DeleteZoneById(gomock.Any(), gomock.Any()).Return(errors.New("DB DOWN")).Times(1)
			}

			zoneService := zone.New(log, saver, provider, deleter)
			r := NewRouter(mux.NewRouter(), zoneService, log)

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/delete/", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})

			r.DeleteZone()(wr, req)
			response := wr.Result()

			require.Equal(t, tt.expectedStatusCode, response.StatusCode)
			require.Equal(t, response.Header.Get("Content-Type"), "application/json")
		})
	}
}
