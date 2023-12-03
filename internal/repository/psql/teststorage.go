package psql

import (
	"context"

	"github.com/maxsnegir/zones_service/internal/config"
	"github.com/maxsnegir/zones_service/internal/db/dbtesting"
	"github.com/maxsnegir/zones_service/internal/logger"
)

type TestStorage struct {
	*Storage
	Resource *dbtesting.PostgresResource
}

func NewTestStorage(ctx context.Context) (*TestStorage, error) {
	log := logger.New(config.EnvTest)
	log.Info("creating docker psql test storage")
	pool, err := dbtesting.ConnectToDocker("")
	if err != nil {
		return nil, err
	}
	log.Info("creating postgres psql resource")
	psqlResource, err := dbtesting.NewPostgresResource(ctx, pool)
	if err != nil {
		return nil, err
	}

	return &TestStorage{
		Storage: &Storage{
			db:  psqlResource.DB,
			log: log,
		},
		Resource: psqlResource,
	}, nil
}

func (t *TestStorage) CleanDB(ctx context.Context) {
	const op = "psql.CleanDB"
	const deleteZoneData = `TRUNCATE zone, zone_geometry RESTART IDENTITY CASCADE;`

	_, err := t.Storage.db.Exec(ctx, deleteZoneData)
	if err != nil {
		t.log.Errorf("%s: failed to clean up test database: %s", op, err.Error())
	}
}

func (t *TestStorage) ShutDown() {
	t.Resource.Close()
}
