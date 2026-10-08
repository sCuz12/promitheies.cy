package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewsletterSubscribeWithoutSectors(t *testing.T) {
	called := false
	s := &Server{storeNewsletterSub: func(_ context.Context, email string, _ Lang, divisions []string, minValue *float64) error {
		called = true
		if email != "test@example.com" {
			t.Errorf("email = %q", email)
		}
		if divisions == nil || len(divisions) != 0 {
			t.Errorf("divisions = %#v, want non-nil empty slice", divisions)
		}
		if minValue != nil {
			t.Errorf("minValue = %v, want nil", minValue)
		}
		return nil
	}}
	r := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", strings.NewReader("email=test%40example.com"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handleNewsletterSubscribe(w, r)

	if !called {
		t.Fatal("subscriber was not stored")
	}
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/?newsletter=joined#weekly-digest" {
		t.Errorf("response = %d, location %q", w.Code, w.Header().Get("Location"))
	}
}
