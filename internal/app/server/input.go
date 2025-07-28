package server

import (
	"fmt"
	"hatchapp/internal/pkg/repository"
)

type TextMessage struct {
	From        string   `json:"from" validate:"required,e164"`                  // E.164 phone number format
	To          string   `json:"to" validate:"required,e164"`                    // E.164 phone number format
	Type        string   `json:"type" validate:"required,oneof=sms mms"`         // Restrict to known types
	Body        string   `json:"body" validate:"required"`                       // Must be non-empty
	Attachments []string `json:"attachments" validate:"omitempty,dive,required"` // Each attachment must be a valid URL if present
	ProviderID  string   `json:"messaging_provider_id"`
	CreatedAt   string   `json:"timestamp" validate:"required,datetime=2006-01-02T15:04:05Z"`
}

func (m *TextMessage) ToRepositoryMessage(status string) (repository.Message, error) {
	msg := repository.Message{
		From:        m.From,
		To:          m.To,
		Type:        m.Type,
		Body:        m.Body,
		Attachments: m.Attachments,
		ProviderID:  m.ProviderID,
		CreatedAt:   m.CreatedAt,
		Status:      status,
	}

	// Determine the communication type based on the message type
	switch m.Type {
	case "sms":
		msg.CommunicationType = repository.CommunicationTypePhone
	case "mms":
		msg.CommunicationType = repository.CommunicationTypePhone
	default:
		return msg, fmt.Errorf("unknown message type: %s", m.Type)
	}

	return msg, nil
}

type EmailMessage struct {
	From        string   `json:"from" validate:"required,email"`                 // Valid email
	To          string   `json:"to" validate:"required,email"`                   // Valid email
	Body        string   `json:"body" validate:"required"`                       // Non-empty body
	Attachments []string `json:"attachments" validate:"omitempty,dive,required"` // Each attachment must be a valid URL if present
	ProviderID  string   `json:"xillio_id"`
	CreatedAt   string   `json:"timestamp" validate:"required,datetime=2006-01-02T15:04:05Z"`
}

func (m *EmailMessage) ToRepositoryMessage(status string) (repository.Message, error) {
	msg := repository.Message{
		From:              m.From,
		To:                m.To,
		Type:              repository.CommunicationTypeEmail,
		CommunicationType: repository.CommunicationTypeEmail,
		Body:              m.Body,
		Attachments:       m.Attachments,
		ProviderID:        m.ProviderID,
		CreatedAt:         m.CreatedAt,
		Status:            status,
	}

	return msg, nil
}
