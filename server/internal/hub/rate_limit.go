package hub

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter is a per-key fixed-window limiter: each key gets maxEvents
// within window, after which further calls are rejected until the window
// rolls over. It exists to stop a single client from spamming room creation
// (each room holds a goroutine + in-memory state) into an unbounded resource
// drain.
type rateLimiter struct {
	mu        sync.Mutex
	maxEvents int
	window    time.Duration
	seen      map[string][]time.Time
}

func newRateLimiter(maxEvents int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		maxEvents: maxEvents,
		window:    window,
		seen:      make(map[string][]time.Time),
	}
}

// Allow reports whether key may proceed now, recording the attempt if so.
func (l *rateLimiter) Allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	events := l.seen[key]
	kept := events[:0]
	for _, t := range events {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= l.maxEvents {
		l.seen[key] = kept
		return false
	}

	l.seen[key] = append(kept, now)
	return true
}

// clientIP extracts the request's real client IP. Production traffic passes
// through Fly.io's proxy, which sets Fly-Client-IP to the actual caller and
// makes r.RemoteAddr the proxy's own address — falling back to RemoteAddr
// alone would put every visitor behind the same rate-limit bucket.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("Fly-Client-IP"); ip != "" {
		return ip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
