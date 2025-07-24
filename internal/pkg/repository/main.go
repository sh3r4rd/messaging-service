package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hatchapp/config"

	"github.com/labstack/gommon/log"
)

var repo Repository

func Initialize(ctx context.Context) error {
	log.Info("Initializing repository...")

	connectionString, found := config.GetValueFromConfig(ctx, "db_connection_string")
	if !found {
		return errors.New("database connection string not found in context")
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	SetRepository(NewRepository(db))
	if err := repo.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Repository initialized successfully")

	return nil
}

func NewRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func SetRepository(r Repository) {
	if r == nil {
		log.Warn("Attempted to set a nil repository, ignoring.")
		return
	}
	repo = r
	log.Info("Repository set successfully.")
}

func GetRepository() (Repository, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository not found, please initialize it first")
	}
	return repo, nil
}
