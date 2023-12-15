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

type Deleter interface {
	DeleteZoneById(ctx context.Context, id int) error
}

type Provider interface {
	GetZonesByIds(ctx context.Context, ids []int) ([]dto.ZoneGeoJSON, error)
	ContainsPoint(ctx context.Context, ids []int, point dto.Point) ([]dto.ZoneContainsPointOut, error)
	AnyContainsPoint(ctx context.Context, ids []int, point dto.Point) (bool, error)
	GetZonesCount(ctx context.Context) (int, error)
	ButchAnyZoneContainsPoint(ctx context.Context, in dto.BatchZoneContainsPointInCollection) ([]dto.BatchZoneContainsPointOut, error)
}

type Service struct {
	log          *logrus.Logger
	zoneSaver    Saver
	zoneProvider Provider
	zoneDeleter  Deleter
}

func New(log *logrus.Logger, zoneSaver Saver, zoneProvider Provider, zoneDeleter Deleter) *Service {
	return &Service{
		log:          log,
		zoneSaver:    zoneSaver,
		zoneProvider: zoneProvider,
		zoneDeleter:  zoneDeleter,
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

func (s *Service) AnyZoneContainsPoint(ctx context.Context, data dto.ZoneContainsPointIn) (bool, error) {
	return s.zoneProvider.AnyContainsPoint(ctx, data.ZoneIds, data.Point)
}

func (s *Service) DeleteZone(ctx context.Context, id int) error {
	return s.zoneDeleter.DeleteZoneById(ctx, id)
}

type BatchZoneContainsPointOutWithError struct {
	dto.BatchZoneContainsPointOut
	Error error
}

func (s *Service) ButchAnyZoneContainsPoint(ctx context.Context, in dto.BatchZoneContainsPointInCollection) ([]dto.BatchZoneContainsPointOut, error) {
	return s.zoneProvider.ButchAnyZoneContainsPoint(ctx, in)
}
