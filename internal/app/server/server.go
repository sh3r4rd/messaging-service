package server

import (
	"encoding/json"
	"hatchapp/internal/pkg/repository"
	"hatchapp/internal/pkg/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/go-playground/validator/v10"
)

type Server struct {
	Repo           repository.Repository
	Validator      *validator.Validate
	MessageService *service.ExternalService
	EmailService   *service.ExternalService
}

// NewServer creates a new instance of the Server with the provided repository.
func NewServer(repo repository.Repository) *Server {
	return &Server{
		Repo:           repo,
		Validator:      validator.New(),
		MessageService: service.NewExternalService(),
		EmailService:   service.NewExternalService(),
	}
}

func (s *Server) Validate(i interface{}) error {
	return s.Validator.Struct(i)
}

func (s *Server) CreateTextMesssage(c echo.Context) error {
	var msg TextMessage

	if err := json.NewDecoder(c.Request().Body).Decode(&msg); err != nil {
		log.Errorf("failed to decode payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
	}

	if err := s.Validate(&msg); err != nil {
		log.Errorf("validation error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	log.Debugf("Received SMS message: %+v\n", msg)
	err := s.MessageService.SendMessageWithRetries(msg.From, msg.To, msg.Body)
	if err != nil {
		log.Errorf("failed to send message: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send message"})
	}

	// Convert to repository message
	repoMsg, err := msg.ToRepositoryMessage()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	msgID, err := s.Repo.CreateMessage(c.Request().Context(), repoMsg)
	if err != nil {
		// TODO: add proper error handling (db errors should not leak to the user)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "received", "message_id": msgID.String()})
}

func (s *Server) GetConversations(c echo.Context) error {
	conversations, err := s.Repo.GetConversations(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, conversations)
}

func (s *Server) GetConversationByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Conversation ID is required"})
	}

	conversation, err := s.Repo.GetConversationByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, conversation)
}
