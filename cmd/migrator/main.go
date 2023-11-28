package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var databaseDsn, migrationsPath, operation string

	flag.StringVar(&databaseDsn, "database-dsn", "", "database dsn")
	flag.StringVar(&migrationsPath, "migrations-path", "../../migrations", "path to migrations")
	flag.StringVar(&operation, "op", "up", "up or down")
	flag.Parse()

	if databaseDsn == "" {
		databaseDsn = os.Getenv("DATABASE_DSN")
	}
	if databaseDsn == "" {
		log.Fatalf("storage-path is required")
	}

	db, err := sql.Open("postgres", databaseDsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("failed to create driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver,
	)

	if err != nil {
		log.Fatalf("failed to create migrator: %s", err)
	}

	switch operation {
	case "up":
		fmt.Println("applying migrations")
		if err := m.Up(); err != nil {
			log.Fatalf("failed to apply migrations: %s", err)
		}
	case "down":
		fmt.Println("down migrations")
		if err := m.Steps(-1); err != nil {
			log.Fatalf("failed to down migrations: %s", err)
		}
	default:
		log.Fatalf("unknown operation: %s", operation)

	}

	fmt.Println("migrations applied")
}
