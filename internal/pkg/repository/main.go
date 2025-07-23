package repository

import (
	"database/sql"
	"fmt"
	"log"
)

var repo Repository

const (
	CommunicationTypeEmail = "email"
	CommunicationTypePhone = "phone"
)

// Message represents the expected JSON payload for SMS messages.
type Message struct {
	From              string   `json:"from"`
	To                string   `json:"to,omitempty"`
	CommunicationType string   `json:"communication_type,omitempty"`
	Type              string   `json:"type"`
	Body              string   `json:"body"`
	Attachments       []string `json:"attachments"`
	ProviderID        string   `json:"provider_id"`
	CreatedAt         string   `json:"timestamp"`
}

// Conversation represents a conversation in the messaging service.
type Conversation struct {
	ID           string           `json:"id"`
	CreatedAt    string           `json:"created_at"`
	Participants []Communications `json:"participants,omitempty"`
	Messages     []Message        `json:"messages,omitempty"`
}

// Communications represents a communication entity.
type Communications struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
}

func Initialize() error {
	log.Println("Initializing repository...")

	db, err := sql.Open("postgres", "postgres://messaging_user:messaging_password@localhost:5432/messaging_service?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	// defer db.Close()
	// TODO: handle db close properly
	repo = NewRepository(db)
	if err := repo.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Repository initialized successfully")

	return nil
}

func NewRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func GetRepository() (Repository, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository not found, please initialize it first")
	}
	return repo, nil
}
