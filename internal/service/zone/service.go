package zone

import (
	"context"
	"log/slog"

	"github.com/maxsnegir/zones_service/internal/domain/dto"
	"github.com/maxsnegir/zones_service/internal/domain/geojson"
)

type Saver interface {
	SaveZoneFromFeatureCollection(ctx context.Context, featureCollection geojson.FeatureCollection) (int, error)
}

type Provider interface {
	GetZonesByIds(ctx context.Context, ids []int) ([]dto.ZoneGeoJSON, error)
}

type Service struct {
	log          *slog.Logger
	zoneSaver    Saver
	zoneProvider Provider
}

func New(log *slog.Logger, zoneSaver Saver, zoneProvider Provider) *Service {
	return &Service{
		log:          log,
		zoneSaver:    zoneSaver,
		zoneProvider: zoneProvider,
	}
}

func (s *Service) SaveZoneFromFeatureCollection(
	ctx context.Context,
	featureCollection geojson.FeatureCollection,
) (int, error) {
	return s.zoneSaver.SaveZoneFromFeatureCollection(ctx, featureCollection)
}

func (s *Service) GetZonesByIds(ctx context.Context, ids []int) ([]dto.ZoneGeoJSON, error) {
	return s.zoneProvider.GetZonesByIds(ctx, ids)
}
