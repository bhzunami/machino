package handler

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"machino/internal/model"

	"golang.org/x/crypto/bcrypt"
)

const minPasswordLen = 8

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func validatePassword(password string) bool {
	return len(strings.TrimSpace(password)) >= minPasswordLen
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if !validateEmail(req.Email) {
		respondError(w, http.StatusBadRequest, "Ungültige E-Mail-Adresse.")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name ist Pflichtfeld.")
		return
	}
	if !validatePassword(req.Password) {
		respondError(w, http.StatusBadRequest, "Passwort muss mindestens 8 Zeichen lang sein.")
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.store.CreateUser(r.Context(), req.Email, req.Name, hash)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("user registered", "email", req.Email, "ip", clientIP(r))
	if !h.createSessionCookie(w, r, user.ID) {
		return
	}
	respond(w, http.StatusCreated, map[string]any{"user": user})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	user, hash, err := h.store.UserByEmail(r.Context(), req.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		h.logger.Warn("login failed", "email", req.Email, "ip", clientIP(r))
		respondError(w, http.StatusUnauthorized, "E-Mail oder Passwort ist falsch.")
		return
	}
	h.logger.Info("login success", "email", req.Email, "user_id", user.ID, "ip", clientIP(r))
	if !h.createSessionCookie(w, r, user.ID) {
		return
	}
	respond(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request, user model.User) {
	if cookie, err := r.Cookie(cookieName); err == nil {
		if err := h.store.DeleteSession(r.Context(), cookie.Value); err != nil {
			h.logger.Warn("delete session", "error", err)
		}
	}
	h.logger.Info("logout", "user_id", user.ID, "ip", clientIP(r))
	http.SetCookie(w, h.sessionCookie(r, "", time.Now().Add(-time.Hour)))
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) me(w http.ResponseWriter, _ *http.Request, user model.User) {
	respond(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	updated, err := h.store.UpdateProfile(r.Context(), user.ID, req.Email, req.Name)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"user": updated})
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.UpdatePassword(r.Context(), user.ID, hash); err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	token, err := h.store.CreatePasswordReset(r.Context(), req.Email, 30*time.Minute)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("password reset requested", "email", req.Email, "ip", clientIP(r))
	if h.mailer.Enabled() {
		if err := h.mailer.SendPasswordReset(req.Email, token); err != nil {
			h.logger.Error("send password reset mail", "error", err)
			respondError(w, http.StatusInternalServerError, "E-Mail konnte nicht gesendet werden.")
			return
		}
		respond(w, http.StatusOK, map[string]string{
			"message": "Reset-Token wurde per E-Mail gesendet.",
		})
		return
	}
	// Demo-Modus: Token direkt zurückgeben wenn kein SMTP konfiguriert.
	respond(w, http.StatusOK, map[string]string{
		"resetToken": token,
		"message":    "Demo-Modus: Kein SMTP konfiguriert, Token wird direkt ausgegeben.",
	})
}

func (h *Handler) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.UsePasswordReset(r.Context(), strings.TrimSpace(req.Token), hash); err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
