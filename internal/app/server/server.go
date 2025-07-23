package server

import (
	"encoding/json"
	"fmt"
	"hatchapp/internal/pkg/repository"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Server struct {
	Repo repository.Repository
}

// NewServer creates a new instance of the Server with the provided repository.
func NewServer(repo repository.Repository) *Server {
	return &Server{
		Repo: repo,
	}
}

func (s *Server) CreateMesssage(c echo.Context) error {
	var msg Message

	if err := json.NewDecoder(c.Request().Body).Decode(&msg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
	}

	fmt.Printf("Received SMS message: %+v\n", msg)

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
