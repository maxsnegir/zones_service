package dbtesting

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	postgresImage    = "mdillon/postgis"
	postgresTag      = "latest"
	postgresUser     = "testuser"
	postgresPassword = "testpassword"
	postgresDbName   = "testdb"
)

type PostgresResource struct {
	DB       *pgxpool.Pool
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func NewPostgresResource(ctx context.Context, pool *dockertest.Pool) (r *PostgresResource, err error) {
	psqlResource, err := RunPostgresContainer(pool)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = pool.Purge(psqlResource)
		}
	}()

	hostAndPort := psqlResource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", postgresUser, postgresPassword, hostAndPort, postgresDbName)
	db, err := ConnectToPostgres(ctx, pool, databaseUrl)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	if err = MakeMigrations(databaseUrl); err != nil {
		return nil, err
	}

	return &PostgresResource{
		DB:       db,
		pool:     pool,
		resource: psqlResource,
	}, nil
}

func (r *PostgresResource) Close() {
	_ = r.pool.Purge(r.resource)
}

func RunPostgresContainer(pool *dockertest.Pool) (*dockertest.Resource, error) {
	const op = "dbtesting.RunPostgresContainer"

	psqlResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: postgresImage,
		Tag:        postgresTag,
		Env: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", postgresPassword),
			fmt.Sprintf("POSTGRES_DB=%s", postgresDbName),
			fmt.Sprintf("POSTGRES_USER=%s", postgresUser),
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return nil, fmt.Errorf("%s: failed to start resource: %w", op, err)
	}
	_ = psqlResource.Expire(120)

	return psqlResource, nil
}

func ConnectToPostgres(ctx context.Context, pool *dockertest.Pool, databaseUrl string) (*pgxpool.Pool, error) {
	const op = "dbtesting.ConnectToPostgres"
	var db *pgxpool.Pool

	err := pool.Retry(func() error {
		var err error
		db, err = pgxpool.Connect(ctx, databaseUrl)
		if err != nil {
			return fmt.Errorf("%s: failed to connect to database: %w", op, err)
		}
		return db.Ping(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	return db, nil
}

func MakeMigrations(dbDsn string) error {
	const migrationsPath = "../../../migrations"
	const op = "dbtesting.MakeMigrations"

	sourcePath := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.New(sourcePath, dbDsn)
	if err != nil {
		return fmt.Errorf("%s: failed to create migrate instance: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("%s: failed to run migrations: %w", op, err)
	}
	return nil
}
