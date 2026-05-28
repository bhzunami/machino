package handler

import (
"net/http"
"strings"
"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
http.ResponseWriter
status      int
wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
if rw.wroteHeader {
return
}
rw.status = code
rw.wroteHeader = true
rw.ResponseWriter.WriteHeader(code)
}

// requestLogger logs every HTTP request with method, path, status, duration and client IP.
// WebSocket upgrade requests are passed through unwrapped so http.Hijacker remains accessible.
func (h *Handler) requestLogger(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
start := time.Now()

// WebSocket upgrades require http.Hijacker — skip wrapping to preserve the interface.
if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
next.ServeHTTP(w, r)
h.logger.Info("http request",
"method", r.Method,
"path", r.URL.Path,
"status", http.StatusSwitchingProtocols,
"duration_ms", time.Since(start).Milliseconds(),
"ip", clientIP(r),
)
return
}

wrapped := wrapResponseWriter(w)
next.ServeHTTP(wrapped, r)

args := []any{
"method", r.Method,
"path", r.URL.Path,
"status", wrapped.status,
"duration_ms", time.Since(start).Milliseconds(),
"ip", clientIP(r),
}
switch {
case wrapped.status >= 500:
h.logger.Error("http request", args...)
case wrapped.status >= 400:
h.logger.Warn("http request", args...)
default:
h.logger.Info("http request", args...)
}
})
}
