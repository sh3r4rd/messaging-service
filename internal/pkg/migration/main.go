package migration

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/labstack/gommon/log"
)

func Initialize(connectionString string) (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return m, nil
}

func Rollback(m *migrate.Migrate, direction string) error {
	version, isDirty, verErr := m.Version()
	if verErr != nil {
		return fmt.Errorf("failed to get dirty migration version: %w", verErr)
	}

	if !isDirty {
		return fmt.Errorf("migration version %d is not dirty, but failed to roll back", version)
	}

	log.Infof("database is dirty at version %d, forcing clean state...\n", version)

	switch direction {
	case "up":
		version--
	case "down":
		version++
	}

	if forceErr := m.Force(int(version)); forceErr != nil {
		return fmt.Errorf("failed to force migration version: %w", forceErr)
	}

	return nil
}
