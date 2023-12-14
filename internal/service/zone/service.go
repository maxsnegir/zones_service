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
	Temp(ctx context.Context, in dto.BatchZoneContainsPointInCollection) ([]dto.BatchZoneContainsPointOut, error)
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
	const op = "service.ButchAnyZoneContainsPoint"
	return s.zoneProvider.Temp(ctx, in)
}

//	var maxWorkers = 50
//
//	workersCnt := len(in)
//	if workersCnt > maxWorkers {
//		workersCnt = maxWorkers
//	}
//	ctx, cancel := context.WithCancel(ctx)
//	defer cancel()
//
//	resultChan := make(chan BatchZoneContainsPointOutWithError, len(in))
//	jobs := make(chan dto.BatchZoneContainsPointIn, len(in))
//	s.log.Errorf("%d", runtime.NumGoroutine())
//	go func() {
//		defer close(jobs)
//
//		for _, v := range in {
//			jobs <- v
//		}
//	}()
//
//	wg := &sync.WaitGroup{}
//	for i := 0; i < workersCnt; i++ {
//		wg.Add(1)
//
//		go func() {
//			defer wg.Done()
//			s.batchAnyZoneWorker(ctx, jobs, resultChan)
//		}()
//	}
//
//	go func() {
//		defer close(resultChan)
//		wg.Wait()
//	}()
//
//	results := make([]dto.BatchZoneContainsPointOut, 0, len(in))
//	for res := range resultChan {
//		if res.Error != nil {
//			return nil, fmt.Errorf("%s: %w", op, res.Error)
//		}
//		results = append(results, res.BatchZoneContainsPointOut)
//	}
//	return results, nil
//}
//
//func (s *Service) batchAnyZoneWorker(
//	ctx context.Context,
//	jobs <-chan dto.BatchZoneContainsPointIn,
//	results chan<- BatchZoneContainsPointOutWithError,
//) {
//
//	for job := range jobs {
//		select {
//		case <-ctx.Done():
//			results <- BatchZoneContainsPointOutWithError{Error: ctx.Err()}
//			return
//		default:
//			contains, err := s.zoneProvider.AnyContainsPoint(ctx, job.ZoneIds, job.Point)
//			results <- BatchZoneContainsPointOutWithError{
//				BatchZoneContainsPointOut: dto.BatchZoneContainsPointOut{
//					Key:      job.Key,
//					Contains: contains,
//				},
//				Error: err,
//			}
//		}
//	}
//}
