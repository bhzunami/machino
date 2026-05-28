package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"machino/internal/mailer"
	"machino/internal/model"
	"machino/internal/realtime"
	"machino/internal/store"

	"github.com/gorilla/mux"
)

const cookieName = "machino_session"

type Handler struct {
	store  *store.Store
	hub    *realtime.Hub
	mailer *mailer.Mailer
	logger *slog.Logger
}

func New(s *store.Store, hub *realtime.Hub, m *mailer.Mailer, logger *slog.Logger) *Handler {
	return &Handler{store: s, hub: hub, mailer: m, logger: logger}
}

func (h *Handler) Router(staticDir string) http.Handler {
r := mux.NewRouter()
api := r.PathPrefix("/api").Subrouter()
api.Use(h.securityHeaders)
api.HandleFunc("/health", h.health).Methods(http.MethodGet)
api.HandleFunc("/auth/register", h.register).Methods(http.MethodPost)
api.HandleFunc("/auth/login", h.login).Methods(http.MethodPost)
api.HandleFunc("/auth/logout", h.auth(h.logout)).Methods(http.MethodPost)
api.HandleFunc("/auth/password-reset/request", h.requestPasswordReset).Methods(http.MethodPost)
api.HandleFunc("/auth/password-reset/confirm", h.confirmPasswordReset).Methods(http.MethodPost)
api.HandleFunc("/me", h.auth(h.me)).Methods(http.MethodGet)
api.HandleFunc("/profile", h.auth(h.updateProfile)).Methods(http.MethodPut)
api.HandleFunc("/profile/password", h.auth(h.updatePassword)).Methods(http.MethodPut)
api.HandleFunc("/projects", h.auth(h.listProjects)).Methods(http.MethodGet)
api.HandleFunc("/projects", h.auth(h.createProject)).Methods(http.MethodPost)
api.HandleFunc("/projects/{projectID}", h.auth(h.updateProject)).Methods(http.MethodPut)
api.HandleFunc("/projects/{projectID}", h.auth(h.deleteProject)).Methods(http.MethodDelete)
api.HandleFunc("/projects/{projectID}/favorite", h.auth(h.setFavorite)).Methods(http.MethodPut)
api.HandleFunc("/projects/{projectID}/todos", h.auth(h.listTodos)).Methods(http.MethodGet)
api.HandleFunc("/projects/{projectID}/todos", h.auth(h.createTodo)).Methods(http.MethodPost)
api.HandleFunc("/projects/{projectID}/todos/completed", h.auth(h.deleteCompletedTodos)).Methods(http.MethodDelete)
api.HandleFunc("/projects/{projectID}/todos/reorder", h.auth(h.reorderTodos)).Methods(http.MethodPut)
api.HandleFunc("/projects/{projectID}/ws", h.auth(h.projectWS)).Methods(http.MethodGet)
api.HandleFunc("/todos/{todoID}", h.auth(h.updateTodo)).Methods(http.MethodPatch)
if staticDir != "" {
r.PathPrefix("/").Handler(spaFileServer(staticDir))
}
return h.securityHeaders(r)
}

type userKey struct{}

func (h *Handler) auth(next func(http.ResponseWriter, *http.Request, model.User)) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
cookie, err := r.Cookie(cookieName)
if err != nil || cookie.Value == "" {
respondError(w, http.StatusUnauthorized, "Bitte anmelden.")
return
}
user, err := h.store.UserBySession(r.Context(), cookie.Value)
if err != nil {
respondError(w, http.StatusUnauthorized, "Bitte erneut anmelden.")
return
}
next(w, r, user)
}
}

func (h *Handler) securityHeaders(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("Referrer-Policy", "same-origin")
w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
next.ServeHTTP(w, r)
})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createSessionCookie(w http.ResponseWriter, r *http.Request, userID string) bool {
token, expiresAt, err := h.store.CreateSession(r.Context(), userID, 14*24*time.Hour)
if err != nil {
h.logger.Error("create session", "error", err)
respondError(w, http.StatusInternalServerError, "Session konnte nicht erstellt werden.")
return false
}
http.SetCookie(w, h.sessionCookie(r, token, expiresAt))
return true
}

func (h *Handler) sessionCookie(r *http.Request, value string, expires time.Time) *http.Cookie {
return &http.Cookie{
Name:     cookieName,
Value:    value,
Path:     "/",
Expires:  expires,
HttpOnly: true,
SameSite: http.SameSiteLaxMode,
Secure:   r.TLS != nil,
}
}

func (h *Handler) handleStoreError(w http.ResponseWriter, err error) {
switch {
case errors.Is(err, store.ErrInvalidInput):
respondError(w, http.StatusBadRequest, "Pflichtfelder fehlen oder sind ungültig.")
case errors.Is(err, store.ErrUnauthorized):
respondError(w, http.StatusUnauthorized, "Nicht erlaubt.")
case errors.Is(err, store.ErrNotFound):
respondError(w, http.StatusNotFound, "Nicht gefunden.")
case errors.Is(err, store.ErrEmailConflict):
respondError(w, http.StatusConflict, "Diese E-Mail wird bereits verwendet.")
default:
h.logger.Error("request failed", "error", err)
respondError(w, http.StatusInternalServerError, "Interner Fehler.")
}
}
