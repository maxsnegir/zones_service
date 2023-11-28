package psql

import (
	"context"
	baseErr "errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pg "github.com/lib/pq"

	"github.com/maxsnegir/zones_service/internal/domain/geojson"
)

type Storage struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, DbConnString string) (*Storage, error) {
	const op = "postgres.New"

	pool, err := pgxpool.Connect(ctx, DbConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log = log.With(slog.String("op", op))
	log.Info("Connected to database")

	return &Storage{
		db:  pool,
		log: log,
	}, nil
}

func (s *Storage) ShutDown() {
	s.db.Close()
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

func (s *Storage) GetZonesByIds(ctx context.Context, ids []int) ([]*geojson.ZoneGEOJSON, error) {
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

	rows, err := s.db.Query(ctx, query, pg.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	result := make([]*geojson.ZoneGEOJSON, 0, len(ids))
	for rows.Next() {
		var zoneGeoJson geojson.ZoneGEOJSON
		err = rows.Scan(&zoneGeoJson.ZoneId, &zoneGeoJson.GeoJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}
		result = append(result, &zoneGeoJson)
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
