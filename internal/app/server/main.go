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

	"hatchapp/config"
	"hatchapp/internal/pkg/repository"
	"hatchapp/internal/pkg/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewServer initializes a new server instance with the provided repository.
func Initialize(server *Server) *echo.Echo {
	e := echo.New()
	e.Validator = server

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	// Routes
	e.POST("/api/messages/sms", server.CreateTextMesssage)
	e.POST("/api/webhooks/sms", server.CreateTextMesssage)
	e.POST("/api/messages/email", server.CreateEmailMessage)
	e.POST("/api/webhooks/email", server.CreateEmailMessage)
	e.GET("/api/conversations", server.GetConversations)
	e.GET("/api/conversations/:id/messages", server.GetConversationByID)

	return e
}

// Run starts the server with the provided context and command.
func Run(ctx context.Context) error {
	repo, err := repository.GetRepository()
	if err != nil {
		return fmt.Errorf("repository not initialized: %w", err)
	}

	sendGridAccountID, found := config.GetValueFromConfig(ctx, "sendgrid_account_sid")
	if !found {
		return errors.New("sendgrid_account_id not found in config")
	}

	sendGridAPIKey, found := config.GetValueFromConfig(ctx, "sendgrid_api_key")
	if !found {
		return errors.New("sendgrid_api_key not found in config")
	}

	twilioAccountSID, found := config.GetValueFromConfig(ctx, "twilio_account_sid")
	if !found {
		return errors.New("twilio_account_sid not found in config")
	}

	twilioAPIKey, found := config.GetValueFromConfig(ctx, "twilio_api_key")
	if !found {
		return errors.New("twilio_api_key not found in config")
	}

	emailService := service.NewEmailService(sendGridAccountID, sendGridAPIKey)
	textService := service.NewTextService(twilioAccountSID, twilioAPIKey)
	server := NewServer(repo, emailService, textService)
	e := Initialize(server)

	go func() {
		err := e.Start(":8080")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Handle graceful shutdown
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
