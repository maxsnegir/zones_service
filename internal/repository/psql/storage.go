package psql

import (
	"context"
	baseErr "errors"
	"fmt"
	"sync"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
	"github.com/maxsnegir/zones_service/internal/dto"
)

type Storage struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func New(ctx context.Context, log *logrus.Logger, DbConnString string) (*Storage, error) {
	const op = "postgres.New"

	pool, err := pgxpool.New(ctx, DbConnString)
	if err != nil {
		return nil, fmt.Errorf("%s:failed to connect to database: %w", op, err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s:failed to ping database: %w", op, err)
	}
	return &Storage{
		db:  pool,
		log: log,
	}, nil
}

func (s *Storage) ShutDown() {
	s.db.Close()
	s.log.Info("Storage shutdown")
}

func (s *Storage) SaveZoneFromFeatureCollection(ctx context.Context, featureCollection geojson.FeatureCollection) (int, error) {
	const createZoneQuery = `INSERT INTO zone DEFAULT VALUES RETURNING id;`
	const createGeometry = `INSERT INTO zone_geometry (zone_id, geom, properties) VALUES ($1, ST_GeomFromEWKB($2), $3)`

	var zoneId int
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return zoneId, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				err = baseErr.Join(err, rollbackErr)
			}
			return
		}
	}()

	err = tx.QueryRow(ctx, createZoneQuery).Scan(&zoneId)
	if err != nil {
		return zoneId, fmt.Errorf("failed to create zone: %w", err)
	}
	for _, feature := range featureCollection.Features {
		_, err = tx.Exec(ctx, createGeometry, zoneId, feature.Geometry.ToEwkb(), feature.Properties)
		if err != nil {
			return zoneId, parsePostgisError(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return zoneId, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return zoneId, nil
}

func (s *Storage) GetZonesByIds(ctx context.Context, ids []int) ([]dto.ZoneGeoJSON, error) {
	const query = `
		SELECT zg.zone_id,
			   jsonb_build_object(
					   'type', 'FeatureCollection',
					   'features', jsonb_agg(
							   jsonb_build_object(
									   'type', 'Feature',
									   'geometry', ST_AsGeoJSON(zg.geom)::jsonb,
									   'properties', zg.properties
							   )
								   )
			   )as geojson
		FROM zone_geometry zg
		WHERE zg.zone_id = any($1)
		GROUP BY zg.zone_id;`

	zoneIds := &pgtype.Int4Array{}
	if err := zoneIds.Set(ids); err != nil {
		return nil, fmt.Errorf("failed to set zone ids: %w", err)
	}

	rows, err := s.db.Query(ctx, query, zoneIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}
	defer rows.Close()
	result := make([]dto.ZoneGeoJSON, 0, len(ids))
	for rows.Next() {
		var zoneGeoJson dto.ZoneGeoJSON
		err = rows.Scan(&zoneGeoJson.ZoneId, &zoneGeoJson.GeoJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}
		result = append(result, zoneGeoJson)
	}
	return result, nil
}

func (s *Storage) GetZonesCount(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM zone;`
	var count int
	err := s.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return count, fmt.Errorf("failed to get zones count: %w", err)
	}

	return count, nil
}

func (s *Storage) ContainsPoint(ctx context.Context, ids []int, point dto.Point) ([]dto.ZoneContainsPointOut, error) {
	const op = "storage.ZonesContainsPoint"
	const query = `
		SELECT zg.zone_id,
			   bool_or(st_contains(zg.geom, st_point($1, $2))) as res
		FROM zone_geometry zg
		WHERE zone_id = any($3)
		GROUP BY zg.zone_id;`

	zoneIds := &pgtype.Int4Array{}
	if err := zoneIds.Set(ids); err != nil {
		return nil, fmt.Errorf("failed to set zone ids: %w", err)
	}

	rows, err := s.db.Query(ctx, query, point.Lon, point.Lat, zoneIds)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check contains point: %w", op, err)
	}
	defer rows.Close()

	result := make([]dto.ZoneContainsPointOut, 0, len(ids))
	for rows.Next() {
		var zoneContainsPointOut dto.ZoneContainsPointOut
		err = rows.Scan(&zoneContainsPointOut.ZoneId, &zoneContainsPointOut.Contains)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan zone: %w", op, err)
		}
		result = append(result, zoneContainsPointOut)
	}

	return result, nil
}

func (s *Storage) AnyContainsPoint(ctx context.Context, ids []int, point dto.Point) (bool, error) {
	con, err := s.db.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to acquire connection: %w", err)
	}

	return s.anyContains(ctx, con, ids, point)
	//const op = "storage.AnyContainsPoint"
	//const query = `
	//	SELECT  CASE WHEN count(*) > 0 THEN true ELSE false END as contains
	//	FROM zone_geometry zg
	//	WHERE zg.zone_id = any($1) and st_contains(zg.geom, st_point($2, $3));`
	//
	//var contains bool
	//zoneIds := &pgtype.Int4Array{}
	//if err := zoneIds.Set(ids); err != nil {
	//	return contains, fmt.Errorf("failed to set zone ids: %w", err)
	//}
	//err := s.db.QueryRow(ctx, query, zoneIds, point.Lon, point.Lat).Scan(&contains)
	//if err != nil {
	//	return contains, fmt.Errorf("%s: failed to check contains point: %w", op, err)
	//}
	//return contains, nil
}

func (s *Storage) DeleteZoneById(ctx context.Context, id int) error {
	const op = "storage.DeleteZoneById"
	const deleteZoneQuery = `DELETE FROM zone WHERE id = $1;`

	_, err := s.db.Exec(ctx, deleteZoneQuery, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete zone: %w", op, err)
	}
	return nil
}

///////

func (s *Storage) anyContains(ctx context.Context, conn *pgxpool.Conn, ids []int, point dto.Point) (bool, error) {
	const op = "storage.AnyContainsPoint"
	const query = `
		SELECT  CASE WHEN count(*) > 0 THEN true ELSE false END as contains
		FROM zone_geometry zg
		WHERE zg.zone_id = any($1) and st_contains(zg.geom, st_point($2, $3));`

	var contains bool
	zoneIds := &pgtype.Int4Array{}
	if err := zoneIds.Set(ids); err != nil {
		return contains, fmt.Errorf("failed to set zone ids: %w", err)
	}
	err := conn.QueryRow(ctx, query, zoneIds, point.Lon, point.Lat).Scan(&contains)
	if err != nil {
		return contains, fmt.Errorf("%s: failed to check contains point: %w", op, err)
	}
	return contains, nil
}

type BatchZoneContainsPointOutWithError struct {
	dto.BatchZoneContainsPointOut
	Error error
}

func (s *Storage) ButchAnyZoneContainsPoint(ctx context.Context, in dto.BatchZoneContainsPointInCollection) ([]dto.BatchZoneContainsPointOut, error) {
	const op = "service.ButchAnyZoneContainsPoint"
	var maxWorkers = 50

	workersCnt := len(in)
	if workersCnt > maxWorkers {
		workersCnt = maxWorkers
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan BatchZoneContainsPointOutWithError, len(in))
	jobs := make(chan dto.BatchZoneContainsPointIn, len(in))

	go func() {
		defer close(jobs)

		for _, v := range in {
			jobs <- v
		}
	}()

	wg := &sync.WaitGroup{}

	for i := 0; i < workersCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			conn, err := s.db.Acquire(ctx)
			if err != nil {
				resultChan <- BatchZoneContainsPointOutWithError{
					Error: fmt.Errorf("%s: failed to acquire connection: %w", op, err),
				}
				return
			}
			defer conn.Release()
			s.batchAnyZoneWorker(ctx, conn, jobs, resultChan)
		}()
	}

	go func() {
		defer close(resultChan)
		wg.Wait()
	}()

	results := make([]dto.BatchZoneContainsPointOut, 0, len(in))
	for res := range resultChan {
		if res.Error != nil {
			return nil, fmt.Errorf("%s: %w", op, res.Error)
		}
		results = append(results, res.BatchZoneContainsPointOut)
	}
	return results, nil
}

func (s *Storage) batchAnyZoneWorker(
	ctx context.Context,
	conn *pgxpool.Conn,
	jobs <-chan dto.BatchZoneContainsPointIn,
	results chan<- BatchZoneContainsPointOutWithError,
) {

	for job := range jobs {
		select {
		case <-ctx.Done():
			results <- BatchZoneContainsPointOutWithError{Error: ctx.Err()}
			return
		default:
			contains, err := s.anyContains(ctx, conn, job.ZoneIds, job.Point)
			results <- BatchZoneContainsPointOutWithError{
				BatchZoneContainsPointOut: dto.BatchZoneContainsPointOut{
					Key:      job.Key,
					Contains: contains,
				},
				Error: err,
			}
		}
	}
}
