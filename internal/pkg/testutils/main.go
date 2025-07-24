package testutils

import (
	"database/sql"
	"hatchapp/internal/app/server"
	"hatchapp/internal/pkg/repository"
	"hatchapp/internal/pkg/service"
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

type ServiceError struct {
	Message string `json:"message"`
}

func NewServer(emailService, textService *service.ExternalService) *echo.Echo {
	repo, _ := repository.GetRepository()
	e := server.Initialize(server.NewServer(repo, emailService, textService))
	return e
}
