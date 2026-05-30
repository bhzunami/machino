package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bhzunami/machino/internal/model"
	"github.com/bhzunami/machino/internal/store"
)

func (h *Handler) setupStatus(w http.ResponseWriter, r *http.Request) {
	required, err := h.store.SetupRequired(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	settings, err := h.store.AppSettings(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"setupRequired": required,
		"settings":      settings,
	})
}

func (h *Handler) completeSetup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email               string `json:"email"`
		Name                string `json:"name"`
		Password            string `json:"password"`
		AppDomain           string `json:"appDomain"`
		RegistrationEnabled bool   `json:"registrationEnabled"`
		SMTPHost            string `json:"smtpHost"`
		SMTPPort            string `json:"smtpPort"`
		SMTPUsername        string `json:"smtpUsername"`
		SMTPPassword        string `json:"smtpPassword"`
		SMTPFrom            string `json:"smtpFrom"`
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
	hash, err := hashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.store.CreateSetupAdmin(r.Context(), req.Email, req.Name, hash, model.AppSettings{
		AppDomain:           req.AppDomain,
		RegistrationEnabled: req.RegistrationEnabled,
		SMTPHost:            req.SMTPHost,
		SMTPPort:            req.SMTPPort,
		SMTPUsername:        req.SMTPUsername,
		SMTPPassword:        req.SMTPPassword,
		SMTPFrom:            req.SMTPFrom,
	})
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			respondError(w, http.StatusConflict, "Setup wurde bereits abgeschlossen.")
			return
		}
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("setup completed", "admin_id", user.ID, "email", user.Email, "ip", clientIP(r))
	if !h.createSessionCookie(w, r, user.ID) {
		return
	}
	respond(w, http.StatusCreated, map[string]any{"user": user})
}
