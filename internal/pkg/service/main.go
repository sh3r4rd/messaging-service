package service

import (
	"fmt"
	"hatchapp/internal/pkg/apperrors"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
)

const MaxRetries = 3

// MockRequest is a mock implementation of an HTTP request for testing purposes.
type MockRequest struct {
	Headers  map[string]string
	Method   string
	URL      string
	Body     string
	Response *http.Response
}

func (r *MockRequest) Do() (*http.Response, error) {
	// Simulate an HTTP request
	fmt.Printf("Mock request: %s %s\n", r.Method, r.URL)
	if r.Body != "" {
		fmt.Printf("Request body: %s\n", r.Body)
	}
	return r.Response, nil
}

type ExternalService struct {
	Request    MockRequest
	RetryCount int
}

func NewEmailService(apiKey, accountID string) *ExternalService {
	return &ExternalService{
		Request: MockRequest{
			Headers: map[string]string{
				"X-API-Key":    apiKey,
				"X-Account-ID": accountID,
			},
			Method: http.MethodPost,
			URL:    "https://api.sendgrid.com/emails",

			Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
			},
		},
	}
}

func NewEmailServiceWithError(apiKey, accountID string, statusCode int, body string) *ExternalService {
	return &ExternalService{
		Request: MockRequest{
			Headers: map[string]string{
				"X-API-Key":    apiKey,
				"X-Account-ID": accountID,
			},
			Method: http.MethodPost,
			URL:    "https://api.sendgrid.com/emails",
			Response: &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(body)),
			},
		},
	}
}

func NewTextService(apiKey, accountID string) *ExternalService {
	return &ExternalService{
		Request: MockRequest{
			Headers: map[string]string{
				"X-API-Key":    apiKey,
				"X-Account-ID": accountID,
			},
			Method: http.MethodPost,
			URL:    "https://api.twilio.com/messages",
			Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
			},
		},
	}
}

func NewTextServiceWithError(apiKey, accountID string, statusCode int, body string) *ExternalService {
	return &ExternalService{
		Request: MockRequest{
			Headers: map[string]string{
				"X-API-Key":    apiKey,
				"X-Account-ID": accountID,
			},
			Method: http.MethodPost,
			URL:    "https://api.twilio.com/messages",
			Response: &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(body)),
			},
		},
	}
}

func (s *ExternalService) SendMessageWithRetries(from, to, body string, attachments []string) (string, error) {
	// Many API services return a 429 Too Many Requests status code when rate limiting.
	// The `Retry-After` header can be used to determine how long to wait before retrying (exponential backoff).
	// I'm opting for a simple retry mechanism here.

	for s.RetryCount < MaxRetries {
		resp, err := s.sendMessage(from, to, body, attachments)
		if err != nil {
			time.Sleep(time.Duration(s.RetryCount+1) * time.Millisecond)
			log.Warnf("Attempt %d failed: %v", s.RetryCount+1, err)
			continue
		}

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return "", apperrors.NewServiceError(err, "unauthorized: check your API key or account ID")
		case http.StatusNotFound:
			return "", apperrors.NewServiceError(err, "not found")
		case http.StatusForbidden:
			return "", apperrors.NewServiceError(err, "access forbidden")
		case http.StatusBadRequest:
			return "", apperrors.NewServiceError(err, "bad request")
		case http.StatusInternalServerError:
			log.Warnf("Retrying due to server error (%d/%d)...\n", s.RetryCount+1, MaxRetries)
			s.RetryCount++
			continue
		case http.StatusTooManyRequests:
			log.Warnf("Rate limit exceeded (%d/%d): %v\n", s.RetryCount+1, MaxRetries, err)
			s.RetryCount++
			continue
		case http.StatusOK:
			log.Infof("Message sent successfully")
			mockID := fmt.Sprintf("message-%d", rand.Intn(100)) // Mock message ID for demonstration
			return mockID, nil
		default:
			log.Warnf("Unexpected status code %d: %v", resp.StatusCode, err)
			return "", apperrors.NewServiceError(err, fmt.Sprintf("unexpected status code from service: %d", resp.StatusCode))
		}
	}

	return "", fmt.Errorf("failed to send message after %d retries", MaxRetries)
}

func (s *ExternalService) sendMessage(from, to, body string, attachments []string) (*http.Response, error) {
	// Simulate sending the message
	fmt.Printf("Sending SMS from %s to %s: %s with attachments [%+v]\n", from, to, body, attachments)

	s.Request.Body = fmt.Sprintf(`{"from":"%s","to":"%s","body":"%s","attachments":%+v}`, from, to, body, attachments)
	resp, err := s.Request.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	return resp, nil
}
