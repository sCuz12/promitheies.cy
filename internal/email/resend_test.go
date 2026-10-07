package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResendClientSendPostsExpectedRequest(t *testing.T) {
	var gotAuth string
	var gotBody resendSendRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewResendClient("test-key", "digest@promitheies.cy")
	client.baseURL = server.URL

	err := client.Send(context.Background(), "user@example.com", "Subject", "<p>hi</p>", "hi")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if want := "Bearer test-key"; gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
	if gotBody.From != "digest@promitheies.cy" {
		t.Errorf("From = %q, want %q", gotBody.From, "digest@promitheies.cy")
	}
	if len(gotBody.To) != 1 || gotBody.To[0] != "user@example.com" {
		t.Errorf("To = %v, want [user@example.com]", gotBody.To)
	}
	if gotBody.Subject != "Subject" || gotBody.HTML != "<p>hi</p>" || gotBody.Text != "hi" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestResendClientSendReturnsErrorOnNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(resendErrorResponse{Name: "validation_error", Message: "invalid to address"})
	}))
	defer server.Close()

	client := NewResendClient("test-key", "digest@promitheies.cy")
	client.baseURL = server.URL

	err := client.Send(context.Background(), "not-an-email", "Subject", "<p>hi</p>", "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
