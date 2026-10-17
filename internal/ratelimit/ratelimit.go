// Package ratelimit provides a per-client-IP token-bucket HTTP middleware.
// State is in memory, which is enough while the app runs on one machine;
// with several machines each enforces its own limit.
package ratelimit

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	evictInterval = 5 * time.Minute
	idleTTL       = 10 * time.Minute
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Limiter struct {
	rps   rate.Limit
	burst int
	ttl   time.Duration
	now   func() time.Time

	mu       sync.Mutex
	visitors map[string]*visitor
}

// New returns a Limiter allowing rps sustained requests per second per IP with
// the given burst. Idle entries are evicted until ctx is cancelled.
func New(ctx context.Context, rps float64, burst int) *Limiter {
	l := &Limiter{
		rps:      rate.Limit(rps),
		burst:    burst,
		ttl:      idleTTL,
		now:      time.Now,
		visitors: make(map[string]*visitor),
	}
	go func() {
		t := time.NewTicker(evictInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				l.evict()
			}
		}
	}()
	return l
}

// Middleware rejects requests over the limit with 429 and Retry-After.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *Limiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = l.now()
	return v.limiter.Allow()
}

func (l *Limiter) evict() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := l.now().Add(-l.ttl)
	for ip, v := range l.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(l.visitors, ip)
		}
	}
}

// clientIP prefers Fly-Client-IP, which Fly's proxy sets on every request, and
// falls back to the socket address (local dev). X-Forwarded-For is deliberately
// ignored: clients can spoof it.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("Fly-Client-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
