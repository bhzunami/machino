package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bhzunami/machino/internal/model"
	"github.com/bhzunami/machino/internal/realtime"
	"github.com/bhzunami/machino/internal/store"

	"github.com/gorilla/mux"
)

const cookieName = "machino_session"

type Handler struct {
	store        *store.Store
	hub          *realtime.Hub
	logger       *slog.Logger
	cookieSecure bool
	authLimiter  *ipRateLimiter
}

func New(s *store.Store, hub *realtime.Hub, logger *slog.Logger) *Handler {
	return &Handler{store: s, hub: hub, logger: logger}
}

func (h *Handler) WithCookieSecure(secure bool) *Handler { h.cookieSecure = secure; return h }

func (h *Handler) Router(staticDir string) http.Handler {
	h.authLimiter = newIPRateLimiter(authRatePerSecond, authBurst)
	rateLimit := h.rateLimitMiddleware(h.authLimiter)

	r := mux.NewRouter()
	r.Use(h.requestLogger)

	api := r.PathPrefix("/api").Subrouter()
	api.Use(h.securityHeaders)

	api.HandleFunc("/health", h.health).Methods(http.MethodGet)

	setupRouter := api.PathPrefix("/setup").Subrouter()
	setupRouter.Use(rateLimit)
	setupRouter.HandleFunc("/status", h.setupStatus).Methods(http.MethodGet)
	setupRouter.HandleFunc("/complete", h.completeSetup).Methods(http.MethodPost)

	// Auth endpoints — rate limited
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.Use(rateLimit)
	authRouter.HandleFunc("/register", h.register).Methods(http.MethodPost)
	authRouter.HandleFunc("/login", h.login).Methods(http.MethodPost)
	authRouter.HandleFunc("/logout", h.auth(h.logout)).Methods(http.MethodPost)
	authRouter.HandleFunc("/password-reset/request", h.requestPasswordReset).Methods(http.MethodPost)
	authRouter.HandleFunc("/password-reset/confirm", h.confirmPasswordReset).Methods(http.MethodPost)

	api.HandleFunc("/me", h.auth(h.me)).Methods(http.MethodGet)
	api.HandleFunc("/profile", h.auth(h.updateProfile)).Methods(http.MethodPut)
	api.HandleFunc("/profile/password", h.auth(h.updatePassword)).Methods(http.MethodPut)
	api.HandleFunc("/users/search", h.auth(h.searchUsers)).Methods(http.MethodGet)
	api.HandleFunc("/admin/users", h.adminAuth(h.listAdminUsers)).Methods(http.MethodGet)
	api.HandleFunc("/admin/users", h.adminAuth(h.createAdminUser)).Methods(http.MethodPost)
	api.HandleFunc("/admin/users/{userID}", h.adminAuth(h.updateAdminUser)).Methods(http.MethodPut)
	api.HandleFunc("/admin/users/{userID}", h.adminAuth(h.deleteAdminUser)).Methods(http.MethodDelete)
	api.HandleFunc("/admin/users/{userID}/password", h.adminAuth(h.updateAdminUserPassword)).Methods(http.MethodPut)
	api.HandleFunc("/admin/settings", h.adminAuth(h.getAdminSettings)).Methods(http.MethodGet)
	api.HandleFunc("/admin/settings", h.adminAuth(h.updateAdminSettings)).Methods(http.MethodPut)
	api.HandleFunc("/projects", h.auth(h.listProjects)).Methods(http.MethodGet)
	api.HandleFunc("/projects", h.auth(h.createProject)).Methods(http.MethodPost)
	api.HandleFunc("/projects/{projectID}", h.auth(h.updateProject)).Methods(http.MethodPut)
	api.HandleFunc("/projects/{projectID}", h.auth(h.deleteProject)).Methods(http.MethodDelete)
	api.HandleFunc("/projects/{projectID}/favorite", h.auth(h.setFavorite)).Methods(http.MethodPut)
	api.HandleFunc("/projects/{projectID}/todos", h.auth(h.listTodos)).Methods(http.MethodGet)
	api.HandleFunc("/projects/{projectID}/todos", h.auth(h.createTodo)).Methods(http.MethodPost)
	api.HandleFunc("/projects/{projectID}/todos/completed", h.auth(h.deleteCompletedTodos)).Methods(http.MethodDelete)
	api.HandleFunc("/projects/{projectID}/todos/reorder", h.auth(h.reorderTodos)).Methods(http.MethodPut)
	api.HandleFunc("/projects/{projectID}/columns", h.auth(h.listColumns)).Methods(http.MethodGet)
	api.HandleFunc("/projects/{projectID}/columns", h.auth(h.createColumn)).Methods(http.MethodPost)
	api.HandleFunc("/projects/{projectID}/columns/reorder", h.auth(h.reorderColumns)).Methods(http.MethodPut)
	api.HandleFunc("/projects/{projectID}/ws", h.auth(h.projectWS)).Methods(http.MethodGet)
	api.HandleFunc("/projects/{projectID}/members", h.auth(h.listMembers)).Methods(http.MethodGet)
	api.HandleFunc("/projects/{projectID}/members", h.auth(h.addMember)).Methods(http.MethodPost)
	api.HandleFunc("/projects/{projectID}/members/{memberUserID}", h.auth(h.removeMember)).Methods(http.MethodDelete)
	api.HandleFunc("/todos/{todoID}", h.auth(h.updateTodo)).Methods(http.MethodPatch)
	api.HandleFunc("/todos/{todoID}", h.auth(h.deleteTodo)).Methods(http.MethodDelete)
	api.HandleFunc("/columns/{columnID}", h.auth(h.updateColumn)).Methods(http.MethodPatch)
	api.HandleFunc("/columns/{columnID}", h.auth(h.deleteColumn)).Methods(http.MethodDelete)

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

func (h *Handler) adminAuth(next func(http.ResponseWriter, *http.Request, model.User)) http.HandlerFunc {
	return h.auth(func(w http.ResponseWriter, r *http.Request, user model.User) {
		if user.Role != model.RoleAdmin {
			respondError(w, http.StatusForbidden, "Admin-Zugriff erforderlich.")
			return
		}
		next(w, r, user)
	})
}

func (h *Handler) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:; font-src 'self'; frame-ancestors 'none'")
		if h.cookieSecure {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) Shutdown() {
	if h.authLimiter != nil {
		h.authLimiter.stop()
	}
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
		Secure:   h.cookieSecure || r.TLS != nil,
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
