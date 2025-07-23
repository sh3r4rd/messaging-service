package server

import (
	"fmt"

	"hatchapp/internal/pkg/repository"

	"github.com/labstack/echo/v4"
)

// Run starts the server with the provided context and command.
func Run() error {

	repo, err := repository.GetRepository()
	if err != nil {
		return fmt.Errorf("repository not initialized: %w", err)
	}

	server := NewServer(repo)

	e := echo.New()
	// Add middleware
	e.POST("/api/messages/sms", server.CreateMesssage)

	fmt.Println("Starting server on :8080...")
	return e.Start(":8080")
}
