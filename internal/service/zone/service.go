package zone

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/dto"
)

type Saver interface {
	SaveZoneFromFeatureCollection(ctx context.Context, featureCollection geojson.FeatureCollection) (int, error)
}

type Provider interface {
	GetZonesByIds(ctx context.Context, ids []int) ([]dto.ZoneGeoJSON, error)
	ContainsPoint(ctx context.Context, ids []int, point dto.Point) ([]dto.ZoneContainsPointOut, error)

	GetZonesCount(ctx context.Context) (int, error)
}

type Service struct {
	log          *logrus.Logger
	zoneSaver    Saver
	zoneProvider Provider
}

func New(log *logrus.Logger, zoneSaver Saver, zoneProvider Provider) *Service {
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

func (s *Service) ContainsPoint(ctx context.Context, data dto.ZoneContainsPointIn) ([]dto.ZoneContainsPointOut, error) {
	return s.zoneProvider.ContainsPoint(ctx, data.ZoneIds, data.Point)
}
