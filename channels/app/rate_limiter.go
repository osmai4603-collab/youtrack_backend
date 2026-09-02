package app

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimitClient struct {
	tokens float64
	seen   time.Time
}

// RateLimiter implements a small token bucket per client identity.
type RateLimiter struct {
	mu        sync.Mutex
	clients   map[string]rateLimitClient
	rate      float64
	burst     float64
	lastSweep time.Time
}

func NewRateLimiter(perMinute, burst int) *RateLimiter {
	if perMinute <= 0 {
		perMinute = 60
	}
	if burst <= 0 {
		burst = 10
	}
	return &RateLimiter{
		clients: make(map[string]rateLimitClient),
		rate:    float64(perMinute) / 60,
		burst:   float64(burst),
	}
}

func (l *RateLimiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	client, ok := l.clients[key]
	if !ok {
		client = rateLimitClient{tokens: l.burst, seen: now}
	}
	client.tokens += now.Sub(client.seen).Seconds() * l.rate
	if client.tokens > l.burst {
		client.tokens = l.burst
	}
	client.seen = now
	if client.tokens < 1 {
		l.clients[key] = client
		return false
	}
	client.tokens--
	l.clients[key] = client

	if now.Sub(l.lastSweep) > time.Minute {
		for clientKey, entry := range l.clients {
			if now.Sub(entry.seen) > 10*time.Minute {
				delete(l.clients, clientKey)
			}
		}
		l.lastSweep = now
	}
	return true
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Authorization")
		if key == "" {
			key = r.RemoteAddr
		}
		key = strings.TrimSpace(key)
		if !l.Allow(key) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
