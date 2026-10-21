package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleOpenAIAppsChallenge(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/.well-known/openai-apps-challenge", nil)
	recorder := httptest.NewRecorder()

	(&Server{}).handleOpenAIAppsChallenge(recorder, req)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if got := result.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", got, "text/plain; charset=utf-8")
	}
	if got := recorder.Body.String(); got != openAIAppsChallengeToken {
		t.Fatalf("body = %q, want exact challenge token", got)
	}
}
