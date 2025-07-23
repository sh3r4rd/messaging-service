package server

import (
	"fmt"
	"hatchapp/internal/pkg/repository"
)

type Message struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	Type        string   `json:"type"`
	Body        string   `json:"body"`
	Attachments []string `json:"attachments"`
	ProviderID  string   `json:"provider_id"`
	CreatedAt   string   `json:"timestamp"`
}

func (m *Message) ToRepositoryMessage() (repository.Message, error) {
	msg := repository.Message{
		From:        m.From,
		To:          m.To,
		Type:        m.Type,
		Body:        m.Body,
		Attachments: m.Attachments,
		ProviderID:  m.ProviderID,
		CreatedAt:   m.CreatedAt,
	}

	// Determine the communication type based on the message type
	switch m.Type {
	case "email":
		msg.CommunicationType = repository.CommunicationTypeEmail
	case "sms":
		msg.CommunicationType = repository.CommunicationTypePhone
	case "mms":
		msg.CommunicationType = repository.CommunicationTypePhone
	default:
		return msg, fmt.Errorf("unknown message type: %s", m.Type)
	}

	return msg, nil
}
