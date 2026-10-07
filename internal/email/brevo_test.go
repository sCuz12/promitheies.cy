package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrevoClientSendPostsExpectedRequest(t *testing.T) {
	var gotAPIKey string
	var gotBody brevoSendRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("api-key")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"messageId": "abc-123"})
	}))
	defer server.Close()

	client := NewBrevoClient("test-key", "digest@promitheies.cy")
	client.baseURL = server.URL

	err := client.Send(context.Background(), "user@example.com", "Subject", "<p>hi</p>", "hi")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotAPIKey != "test-key" {
		t.Errorf("api-key header = %q, want %q", gotAPIKey, "test-key")
	}
	if gotBody.Sender.Email != "digest@promitheies.cy" {
		t.Errorf("Sender = %q, want %q", gotBody.Sender.Email, "digest@promitheies.cy")
	}
	if len(gotBody.To) != 1 || gotBody.To[0].Email != "user@example.com" {
		t.Errorf("To = %v, want [user@example.com]", gotBody.To)
	}
	if gotBody.Subject != "Subject" || gotBody.HTMLContent != "<p>hi</p>" || gotBody.TextContent != "hi" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestBrevoClientSendReturnsErrorOnNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(brevoErrorResponse{Code: "invalid_parameter", Message: "invalid 'to' email address"})
	}))
	defer server.Close()

	client := NewBrevoClient("test-key", "digest@promitheies.cy")
	client.baseURL = server.URL

	err := client.Send(context.Background(), "not-an-email", "Subject", "<p>hi</p>", "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
