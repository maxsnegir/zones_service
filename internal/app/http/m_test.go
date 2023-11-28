package http

import (
	"context"
	baseLog "log"
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/maxsnegir/zones_service/internal/config"
	"github.com/maxsnegir/zones_service/internal/logger"
	"github.com/maxsnegir/zones_service/internal/repository/psql"
)

var (
	storage *psql.TestStorage
	log     = logger.New(config.EnvTest)
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	testStorage, err := psql.NewTestStorage(ctx)
	if err != nil {
		baseLog.Fatal(err)
	}
	storage = testStorage
	code := m.Run()
	storage.ShutDown()

	os.Exit(code)
}
