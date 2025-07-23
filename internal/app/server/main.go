package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hatchapp/internal/pkg/repository"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Run starts the server with the provided context and command.
func Run() error {

	repo, err := repository.GetRepository()
	if err != nil {
		return fmt.Errorf("repository not initialized: %w", err)
	}

	server := NewServer(repo)

	e := echo.New()
	e.Validator = server

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	// Routes
	e.POST("/api/messages/sms", server.CreateTextMesssage)
	e.GET("/api/conversations", server.GetConversations)
	e.GET("/api/conversations/:id/messages", server.GetConversationByID)

	go func() {
		err := e.Start(":8080")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit
	log.Println("Shutdown signal received, starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}

	log.Println("Server exited gracefully")
	return nil
}
