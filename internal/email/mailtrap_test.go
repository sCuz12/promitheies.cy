package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestMailtrapClientSendPostsExpectedRequest(t *testing.T) {
	var gotAuth string
	var gotBody mailtrap.SendRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message_ids":["abc-123"]}`))
	}))
	defer server.Close()

	client, err := NewMailtrapClient("test-token", "digest@promitheies.cy", mailtrap.WithBaseURL(mailtrap.HostSend, server.URL))
	if err != nil {
		t.Fatalf("NewMailtrapClient returned error: %v", err)
	}

	err = client.Send(context.Background(), "user@example.com", "Subject", "<p>hi</p>", "hi")
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if want := "Bearer test-token"; gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
	if gotBody.From.Email != "digest@promitheies.cy" {
		t.Errorf("From = %q, want %q", gotBody.From.Email, "digest@promitheies.cy")
	}
	if len(gotBody.To) != 1 || gotBody.To[0].Email != "user@example.com" {
		t.Errorf("To = %v, want [user@example.com]", gotBody.To)
	}
	if gotBody.Subject != "Subject" || gotBody.HTML != "<p>hi</p>" || gotBody.Text != "hi" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestMailtrapClientSendReturnsErrorOnUnsuccessfulResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"errors":["'to' is invalid"]}`))
	}))
	defer server.Close()

	client, err := NewMailtrapClient("test-token", "digest@promitheies.cy", mailtrap.WithBaseURL(mailtrap.HostSend, server.URL))
	if err != nil {
		t.Fatalf("NewMailtrapClient returned error: %v", err)
	}

	err = client.Send(context.Background(), "not-an-email", "Subject", "<p>hi</p>", "hi")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
