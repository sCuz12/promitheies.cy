package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLimiter(rps float64, burst int) *Limiter {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // stop the evictor; tests call evict directly
	return New(ctx, rps, burst)
}

func do(h http.Handler, remote string, hdr map[string]string) int {
	req := httptest.NewRequest("GET", "/mcp", nil)
	req.RemoteAddr = remote
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

var ok = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

func TestBurstThenRejected(t *testing.T) {
	h := newTestLimiter(0.001, 3).Middleware(ok)
	for i := range 3 {
		if got := do(h, "1.1.1.1:1000", nil); got != 200 {
			t.Fatalf("request %d = %d, want 200", i, got)
		}
	}
	if got := do(h, "1.1.1.1:1000", nil); got != http.StatusTooManyRequests {
		t.Fatalf("over burst = %d, want 429", got)
	}
}

func TestIPsAreIndependent(t *testing.T) {
	h := newTestLimiter(0.001, 1).Middleware(ok)
	do(h, "1.1.1.1:1", nil)
	if got := do(h, "2.2.2.2:1", nil); got != 200 {
		t.Fatalf("second IP = %d, want 200", got)
	}
}

func TestFlyClientIPHonouredAndXFFIgnored(t *testing.T) {
	h := newTestLimiter(0.001, 1).Middleware(ok)
	do(h, "10.0.0.1:1", map[string]string{"Fly-Client-IP": "9.9.9.9"})
	// Same Fly client IP from another socket: still limited.
	if got := do(h, "10.0.0.2:1", map[string]string{"Fly-Client-IP": "9.9.9.9"}); got != 429 {
		t.Fatalf("same Fly-Client-IP = %d, want 429", got)
	}
	// Spoofed X-Forwarded-For must not reset the bucket.
	if got := do(h, "10.0.0.1:1", map[string]string{"X-Forwarded-For": "8.8.8.8"}); got != 200 {
		// first request from socket 10.0.0.1 without Fly header is a new key
		t.Fatalf("socket IP fresh bucket = %d, want 200", got)
	}
	if got := do(h, "10.0.0.1:1", map[string]string{"X-Forwarded-For": "7.7.7.7"}); got != 429 {
		t.Fatalf("spoofed XFF = %d, want 429", got)
	}
}

func TestEvictRemovesIdle(t *testing.T) {
	l := newTestLimiter(1, 1)
	now := time.Now()
	l.now = func() time.Time { return now }
	l.allow("1.1.1.1")
	l.now = func() time.Time { return now.Add(l.ttl + time.Second) }
	l.allow("2.2.2.2")
	l.evict()
	if _, ok := l.visitors["1.1.1.1"]; ok {
		t.Fatal("idle visitor not evicted")
	}
	if _, ok := l.visitors["2.2.2.2"]; !ok {
		t.Fatal("active visitor evicted")
	}
}
