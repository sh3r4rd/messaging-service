package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"hatchapp/internal/pkg/apperrors"
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
func NewServer(repo repository.Repository, emailService, textService *service.ExternalService) *Server {
	return &Server{
		Repo:           repo,
		Validator:      validator.New(),
		MessageService: textService,
		EmailService:   emailService,
	}
}

func (s *Server) Validate(i interface{}) error {
	return s.Validator.Struct(i)
}

func (s *Server) CreateTextMesssage(c echo.Context) error {
	var msg TextMessage

	if err := json.NewDecoder(c.Request().Body).Decode(&msg); err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusUnprocessableEntity, "invalid payload: failed to decode json")
	}

	if err := s.Validate(&msg); err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusUnprocessableEntity, "invalid request input")
	}

	// If the request is for the SMS endpoint, send the message via the external service.
	// If the request is for the webhook endpoint, we assume the message has already been sent by the provider.
	if path := c.Path(); path == "/api/messages/sms" {
		err := s.MessageService.SendMessageWithRetries(msg.From, msg.To, msg.Body, msg.Attachments)
		if err != nil {
			return apperrors.ApiErrorResponse(c, err, http.StatusInternalServerError, "failed to send message via provider")
		}
	}

	// Convert to repository message
	repoMsg, err := msg.ToRepositoryMessage()
	if err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusUnprocessableEntity, "failed to convert message")
	}

	msgID, err := s.Repo.CreateMessage(c.Request().Context(), repoMsg)
	if err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusInternalServerError, "failed to store text message")
	}

	return c.JSON(http.StatusCreated, map[string]string{"status": "received", "message_id": fmt.Sprintf("%d", *msgID)})
}

func (s *Server) CreateEmailMessage(c echo.Context) error {
	var emailMsg EmailMessage

	if err := json.NewDecoder(c.Request().Body).Decode(&emailMsg); err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusUnprocessableEntity, "invalid payload: failed to decode json")
	}

	if err := s.Validate(&emailMsg); err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusUnprocessableEntity, "invalid request input")
	}

	// If the request is for the email endpoint, send the message via the external service.
	if path := c.Path(); path == "/api/messages/email" {
		err := s.EmailService.SendMessageWithRetries(emailMsg.From, emailMsg.To, emailMsg.Body, emailMsg.Attachments)
		if err != nil {
			return apperrors.ApiErrorResponse(c, err, http.StatusInternalServerError, "failed to send email via provider")
		}
	}

	// Convert to repository message
	repoEmailMsg, err := emailMsg.ToRepositoryMessage()
	if err != nil {
		log.Errorf("failed to convert email message: %v", err)
		return apperrors.ApiErrorResponse(c, err, http.StatusUnprocessableEntity, "failed to convert email message")
	}

	msgID, err := s.Repo.CreateMessage(c.Request().Context(), repoEmailMsg)
	if err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusInternalServerError, "failed to store email message")
	}

	return c.JSON(http.StatusCreated, map[string]string{"status": "received", "message_id": fmt.Sprintf("%d", *msgID)})
}

func (s *Server) GetConversations(c echo.Context) error {
	conversations, err := s.Repo.GetConversations(c.Request().Context())
	if err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusInternalServerError, "failed to get conversations")
	}

	return c.JSON(http.StatusOK, conversations)
}

func (s *Server) GetConversationByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		err := errors.New("conversation ID is required")
		return apperrors.ApiErrorResponse(c, err, http.StatusBadRequest, "conversation ID is required")
	}

	conversation, err := s.Repo.GetConversationByID(c.Request().Context(), id)
	if err != nil {
		return apperrors.ApiErrorResponse(c, err, http.StatusInternalServerError, "failed to get conversation")
	}

	return c.JSON(http.StatusOK, conversation)
}
