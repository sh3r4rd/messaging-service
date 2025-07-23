package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const MaxRetries = 3

// MockRequest is a mock implementation of an HTTP request for testing purposes.
type MockRequest struct {
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
	retryCount int
}

func NewExternalService() *ExternalService {
	return &ExternalService{
		Request: MockRequest{
			Method: http.MethodPost,
			URL:    "https://api.example.com/messages",
			Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success"}`)),
			},
		},
	}
}

func NewExternalServiceWithError(statusCode int, body string) *ExternalService {
	return &ExternalService{
		Request: MockRequest{
			Method: http.MethodPost,
			URL:    "https://api.example.com/messages",
			Response: &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(body)),
			},
		},
	}
}

func (s *ExternalService) SendMessageWithRetries(from, to, body string) error {
	// Implement retry logic here with exponential backoff
	for s.retryCount < MaxRetries {
		resp, err := s.sendMessage(from, to, body)
		if err != nil {
			time.Sleep(time.Duration(s.retryCount) * time.Second) // Exponential backoff
			log.Printf("Attempt %d failed: %v", s.retryCount+1, err)
			continue
		}

		switch resp.StatusCode {
		case http.StatusInternalServerError:
			fmt.Printf("Retrying due to server error (%d/%d)...\n", s.retryCount+1, MaxRetries)
			continue
		case http.StatusBadRequest:
			fmt.Printf("Bad request (%d/%d): %v\n", s.retryCount+1, MaxRetries, err)
			continue
		case http.StatusTooManyRequests:
			fmt.Printf("Rate limit exceeded (%d/%d): %v\n", s.retryCount+1, MaxRetries, err)
			continue
		case http.StatusOK:
			fmt.Println("Message sent successfully")
			return nil
		}

		s.retryCount++
	}

	return fmt.Errorf("failed to send message after %d retries", MaxRetries)
}

func (s *ExternalService) sendMessage(from, to, body string) (*http.Response, error) {
	// Simulate sending the message
	fmt.Printf("Sending SMS from %s to %s: %s\n", from, to, body)

	s.Request.Body = fmt.Sprintf(`{"from":"%s","to":"%s","body":"%s"}`, from, to, body)
	resp, err := s.Request.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	return resp, nil
}
