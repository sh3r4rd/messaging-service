package repository

const (
	CommunicationTypeEmail = "email"
	CommunicationTypePhone = "phone"
)

const (
	MessageStatusSuccess = "success"
	MessageStatusFailed  = "failed"
)

// Message represents the expected JSON payload for SMS messages.
type Message struct {
	ID                int64    `json:"id,omitempty"`
	From              string   `json:"from"`
	To                string   `json:"to,omitempty"`
	CommunicationType string   `json:"communication_type,omitempty"`
	Type              string   `json:"type"`
	Body              string   `json:"body"`
	Attachments       []string `json:"attachments"`
	ProviderID        string   `json:"provider_id"`
	Status            string   `json:"status,omitempty"`
	CreatedAt         string   `json:"timestamp"`
}

// Conversation represents a conversation in the messaging service.
type Conversation struct {
	ID           int64           `json:"id"`
	CreatedAt    string          `json:"created_at"`
	Participants []Communication `json:"participants,omitempty"`
	Messages     []Message       `json:"messages,omitempty"`
}

// Communications represents a communication entity.
type Communication struct {
	ID         int64  `json:"id"`
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
}
