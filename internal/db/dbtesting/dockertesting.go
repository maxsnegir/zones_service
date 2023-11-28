package dbtesting

import (
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
)

func ConnectToDocker(endpoint string) (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool(endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}
	err = pool.Client.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 1200 * time.Second
	return pool, nil
}
