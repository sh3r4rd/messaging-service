package repository

import (
	"database/sql"
	"fmt"

	"github.com/labstack/gommon/log"
)

var repo Repository

func Initialize() error {
	log.Info("Initializing repository...")

	db, err := sql.Open("postgres", "postgres://messaging_user:messaging_password@localhost:5432/messaging_service?sslmode=disable")
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
