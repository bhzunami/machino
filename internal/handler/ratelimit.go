package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	authRatePerSecond = rate.Limit(1.0 / 6.0) // 1 request per 6 seconds
	authBurst         = 5
	cleanupInterval   = 5 * time.Minute
	entryTTL          = 15 * time.Minute
)

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	r       rate.Limit
	burst   int
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{
		entries: make(map[string]*ipEntry),
		r:       r,
		burst:   burst,
	}
	go l.cleanup()
	return l
}

func (l *ipRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.entries[ip]; ok {
		e.lastSeen = time.Now()
		return e.limiter
	}
	limiter := rate.NewLimiter(l.r, l.burst)
	l.entries[ip] = &ipEntry{limiter: limiter, lastSeen: time.Now()}
	return limiter
}

func (l *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for ip, e := range l.entries {
			if time.Since(e.lastSeen) > entryTTL {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}

// rateLimitMiddleware returns a middleware that limits requests per client IP.
func (h *Handler) rateLimitMiddleware(limiter *ipRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !limiter.get(ip).Allow() {
				h.logger.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
				respondError(w, http.StatusTooManyRequests, "Zu viele Anfragen. Bitte warte kurz.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the real client IP, respecting X-Forwarded-For from trusted proxies.
// Note: Only trust X-Forwarded-For if running behind a known reverse proxy.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain — the original client
		if ip := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
