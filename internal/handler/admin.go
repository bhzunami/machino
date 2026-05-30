package handler

import (
	"net/http"
	"strings"

	"github.com/bhzunami/machino/internal/model"

	"github.com/gorilla/mux"
)

func (h *Handler) listAdminUsers(w http.ResponseWriter, r *http.Request, admin model.User) {
	users, err := h.store.ListUsers(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"users": users})
}

func (h *Handler) createAdminUser(w http.ResponseWriter, r *http.Request, admin model.User) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.Role = strings.TrimSpace(req.Role)
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
	if req.Role != model.RoleUser && req.Role != model.RoleAdmin {
		req.Role = model.RoleUser
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.store.CreateUser(r.Context(), req.Email, req.Name, hash, req.Role)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("admin created user", "admin_id", admin.ID, "new_user_id", created.ID)
	respond(w, http.StatusOK, map[string]any{"user": created})
}

func (h *Handler) updateAdminUser(w http.ResponseWriter, r *http.Request, admin model.User) {
	userID := mux.Vars(r)["userID"]
	var req struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		Searchable bool   `json:"searchable"`
		Role       string `json:"role"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.Role = strings.TrimSpace(req.Role)
	if !validateEmail(req.Email) {
		respondError(w, http.StatusBadRequest, "Ungültige E-Mail-Adresse.")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name ist Pflichtfeld.")
		return
	}
	if req.Role != model.RoleUser && req.Role != model.RoleAdmin {
		respondError(w, http.StatusBadRequest, "Ungültige Rolle.")
		return
	}
	if userID == admin.ID && req.Role != model.RoleAdmin {
		respondError(w, http.StatusBadRequest, "Du kannst dir die Admin-Rolle nicht selbst entziehen.")
		return
	}
	updated, err := h.store.UpdateUser(r.Context(), userID, req.Email, req.Name, req.Searchable, req.Role)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("admin user updated", "admin_id", admin.ID, "target_user_id", userID)
	respond(w, http.StatusOK, map[string]any{"user": updated})
}

func (h *Handler) deleteAdminUser(w http.ResponseWriter, r *http.Request, admin model.User) {
	userID := mux.Vars(r)["userID"]
	if userID == admin.ID {
		respondError(w, http.StatusBadRequest, "Du kannst deinen eigenen Admin-User nicht löschen.")
		return
	}
	if err := h.store.DeleteUser(r.Context(), userID); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("admin user deleted", "admin_id", admin.ID, "target_user_id", userID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) updateAdminUserPassword(w http.ResponseWriter, r *http.Request, admin model.User) {
	userID := mux.Vars(r)["userID"]
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
	if err := h.store.UpdatePassword(r.Context(), userID, hash); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("admin user password reset", "admin_id", admin.ID, "target_user_id", userID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getAdminSettings(w http.ResponseWriter, r *http.Request, admin model.User) {
	settings, err := h.store.AppSettings(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"settings": settings})
}

func (h *Handler) updateAdminSettings(w http.ResponseWriter, r *http.Request, admin model.User) {
	var req struct {
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
	settings, err := h.store.UpdateAppSettings(r.Context(), model.AppSettings{
		AppDomain:           req.AppDomain,
		RegistrationEnabled: req.RegistrationEnabled,
		SMTPHost:            req.SMTPHost,
		SMTPPort:            req.SMTPPort,
		SMTPUsername:        req.SMTPUsername,
		SMTPPassword:        req.SMTPPassword,
		SMTPFrom:            req.SMTPFrom,
	})
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.logger.Info("admin settings updated", "admin_id", admin.ID)
	respond(w, http.StatusOK, map[string]any{"settings": settings})
}
