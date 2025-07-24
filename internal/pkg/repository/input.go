package repository

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
	ID           int64            `json:"id"`
	CreatedAt    string           `json:"created_at"`
	Participants []Communications `json:"participants,omitempty"`
	Messages     []Message        `json:"messages,omitempty"`
}

// Communications represents a communication entity.
type Communications struct {
	ID         int64  `json:"id"`
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
}
