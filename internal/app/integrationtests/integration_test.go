package integrationtests_test

import (
	"fmt"
	"hatchapp/internal/app/server"
	"hatchapp/internal/pkg/repository"
	"hatchapp/internal/pkg/testutils"
	"net/http"
	"os"
	"testing"

	oapi "github.com/oapi-codegen/testutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"
)

var tables = []string{
	"messages",
	"conversations",
	"conversation_memberships",
	"communications",
}

func TestMain(m *testing.M) {
	testutils.SetupTestEnvironment()
	code := m.Run()
	testutils.TeardownTestEnvironment()
	os.Exit(code)
}

func TestMessagesAndConversations(t *testing.T) {
	postgres := engine.NewPostgresEngine(testutils.ConnectionString)
	cleaner := dbcleaner.New()
	cleaner.SetEngine(postgres)

	e := testutils.NewServer()

	t.Run("send and save SMS message", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)

		path := "/api/messages/sms"
		body := server.TextMessage{
			From:        "+1234567890",
			To:          "+0987654321",
			Type:        "sms",
			Body:        "Hello, this is a test message.",
			Attachments: []string{},
			CreatedAt:   "2023-10-01T12:00:00Z",
		}

		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}

		result := make(map[string]string)
		if err := response.UnmarshalBodyToObject(&result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "received", result["status"])
		assert.NotEmpty(t, result["message_id"], "Expected message_id to be present in response")
	})

	t.Run("send and save MMS message", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)

		path := "/api/messages/sms"
		body := server.TextMessage{
			From:        "+1234567890",
			To:          "+0987654321",
			Type:        "mms",
			Body:        "Hello, this is a test message.",
			Attachments: []string{},
			CreatedAt:   "2023-10-01T12:00:00Z",
		}

		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}

		result := make(map[string]string)
		if err := response.UnmarshalBodyToObject(&result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "received", result["status"])
		assert.NotEmpty(t, result["message_id"], "Expected message_id to be present in response")
	})

	t.Run("incoming SMS via webhook", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)
		path := "/api/webhooks/sms"
		body := server.TextMessage{
			From:        "+1234567890",
			To:          "+0987654321",
			Type:        "sms",
			Body:        "Hello, this is a test message via webhook.",
			Attachments: []string{"http://example.com/image.jpg"},
			ProviderID:  "provider123",
			CreatedAt:   "2023-10-01T12:00:00Z",
		}
		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}
		result := make(map[string]string)
		if err := response.UnmarshalBodyToObject(&result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "received", result["status"])
		assert.NotEmpty(t, result["message_id"], "Expected message_id to be present in response")
	})

	t.Run("incoming MMS via webhook", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)
		path := "/api/webhooks/sms"
		body := server.TextMessage{
			From:        "+1234567890",
			To:          "+0987654321",
			Type:        "mms",
			Body:        "Hello, this is a test message via webhook.",
			Attachments: []string{"http://example.com/image.jpg"},
			ProviderID:  "provider123",
			CreatedAt:   "2023-10-01T12:00:00Z",
		}
		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}
		result := make(map[string]string)
		if err := response.UnmarshalBodyToObject(&result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "received", result["status"])
		assert.NotEmpty(t, result["message_id"], "Expected message_id to be present in response")
	})

	t.Run("send and save email message", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)
		path := "/api/messages/email"
		body := server.EmailMessage{
			From:        "sender@example.com",
			To:          "recipient@example.com",
			Body:        "Hello, this is a test email.",
			Attachments: []string{},
			CreatedAt:   "2023-10-01T12:00:00Z",
		}

		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}

		result := make(map[string]string)
		if err := response.UnmarshalBodyToObject(&result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "received", result["status"])
		assert.NotEmpty(t, result["message_id"], "Expected message_id to be present in response")
	})

	t.Run("incoming email via webhook", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)

		path := "/api/webhooks/email"
		body := server.EmailMessage{
			From:        "sender@example.com",
			To:          "recipient@example.com",
			Body:        "Hello, this is a test email.",
			Attachments: []string{},
			CreatedAt:   "2023-10-01T12:00:00Z",
		}

		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}

		result := make(map[string]string)
		if err := response.UnmarshalBodyToObject(&result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		assert.Equal(t, "received", result["status"])
		assert.NotEmpty(t, result["message_id"], "Expected message_id to be present in response")
	})

	t.Run("get conversations", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)

		// Send a message to create a conversation
		path := "/api/messages/sms"
		body := server.TextMessage{
			From:        "+1234567890",
			To:          "+0987654321",
			Type:        "sms",
			Body:        "Hello, this is a test message.",
			Attachments: []string{},
			ProviderID:  "provider123",
			CreatedAt:   "2023-10-01T12:00:00Z",
		}

		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}

		path = "/api/conversations"
		response = oapi.NewRequest().Get(path).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", response.Code())
		}

		var conversations []repository.Conversation
		if err := response.UnmarshalBodyToObject(&conversations); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Len(t, conversations, 1, "Expected one conversation to be returned")
	})

	t.Run("get conversation by ID", func(t *testing.T) {
		cleaner.Acquire(tables...)
		defer cleaner.Clean(tables...)

		// Send a message to create a conversation
		path := "/api/messages/sms"
		body := server.TextMessage{
			From:        "+1234567890",
			To:          "+0987654321",
			Type:        "sms",
			Body:        "Hello, this is a test message.",
			Attachments: []string{},
			ProviderID:  "provider123",
			CreatedAt:   "2023-10-01T12:00:00Z",
		}

		response := oapi.NewRequest().WithHeader("Content-Type", "application/json").Post(path).WithJsonBody(body).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusCreated {
			t.Fatalf("Expected status code 201, got %d", response.Code())
		}

		// Get the conversation ID from the response
		path = "/api/conversations"
		response = oapi.NewRequest().Get(path).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", response.Code())
		}

		var conversations []repository.Conversation
		if err := response.UnmarshalBodyToObject(&conversations); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Code() != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", response.Code())
		}

		assert.Len(t, conversations, 1, "Expected one conversation to be returned")

		// Get the conversation by ID
		conversationID := conversations[0].ID
		path = fmt.Sprintf("/api/conversations/%d/messages", conversationID)
		response = oapi.NewRequest().Get(path).GoWithHTTPHandler(t, e)
		if response.Code() != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", response.Code())
		}

		var conversation repository.Conversation
		if err := response.UnmarshalBodyToObject(&conversation); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.NotEmpty(t, conversation.ID, "expected conversation ID to be present")
		assert.Len(t, conversation.Messages, 1, "Expected one message in the conversation")
	})
}
