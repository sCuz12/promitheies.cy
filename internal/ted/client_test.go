package ted

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(url string) *Client {
	c := NewClient(url)
	c.backoff = func(int) time.Duration { return time.Millisecond }
	return c
}

func TestSearchRetriesForbiddenThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != userAgent {
			t.Errorf("User-Agent = %q", got)
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("<html>blocked</html>"))
			return
		}
		_, _ = w.Write([]byte(`{"notices":[],"totalNoticeCount":0}`))
	}))
	defer srv.Close()

	if _, err := testClient(srv.URL).Search(context.Background(), "q", nil, 1, 10); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestSearchPersistentForbiddenReportsDiagnostics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Amz-Cf-Id", "abc123")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("<html>The request could not be satisfied</html>"))
	}))
	defer srv.Close()

	_, err := testClient(srv.URL).Search(context.Background(), "q", nil, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"http 403", "X-Amz-Cf-Id=abc123", "could not be satisfied"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
}

func TestSearchDoesNotRetryClientError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"bad query"}`))
	}))
	defer srv.Close()

	if _, err := testClient(srv.URL).Search(context.Background(), "q", nil, 1, 10); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestQuerySince(t *testing.T) {
	now := time.Date(2026, 11, 12, 8, 0, 0, 0, time.UTC)
	if got, want := QuerySince(now, 14), "place-of-performance=CYP AND publication-date>=20261029"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := QuerySince(now, 0); got != CyprusQuery {
		t.Errorf("lookback 0 should be unfiltered, got %q", got)
	}
}
