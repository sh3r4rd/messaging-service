package testutils

import (
	"database/sql"
	"hatchapp/internal/app/server"
	"hatchapp/internal/pkg/repository"
	"log"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var ConnectionString = "postgres://messaging_user:messaging_password@localhost:5434/messaging_service_test?sslmode=disable"

// SetupTestEnvironment initializes the test environment.
func SetupTestEnvironment() {
	db, err := sql.Open("postgres", ConnectionString)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	repository.SetRepository(repository.NewRepository(db))
}

// TeardownTestEnvironment cleans up the test environment.
func TeardownTestEnvironment() {
	// Clean up any resources used during tests.
	repo, err := repository.GetRepository()
	if err != nil {
		log.Fatalf("failed to get repository: %s", err)
	}

	repo.Close()
}

func NewServer() *echo.Echo {
	repo, _ := repository.GetRepository()
	e := server.Initialize(server.NewServer(repo))
	return e
}
