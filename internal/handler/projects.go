package handler

import (
	"net/http"

	"github.com/bhzunami/machino/internal/model"

	"github.com/gorilla/mux"
)

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request, user model.User) {
	projects, err := h.store.ListProjects(r.Context(), user.ID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"projects": projects})
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Color       string `json:"color"`
		MoveDone    *bool  `json:"moveDone"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	moveDone := true
	if req.MoveDone != nil {
		moveDone = *req.MoveDone
	}
	project, err := h.store.CreateProject(r.Context(), user.ID, req.Title, req.Description, req.Color, moveDone)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusCreated, map[string]any{"project": project})
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Color       string `json:"color"`
		MoveDone    *bool  `json:"moveDone"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	moveDone := true
	if req.MoveDone != nil {
		moveDone = *req.MoveDone
	}
	project, err := h.store.UpdateProject(r.Context(), user.ID, projectID, req.Title, req.Description, req.Color, moveDone)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"project": project})
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.DeleteProject(r.Context(), user.ID, projectID); err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) setFavorite(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Favorite bool `json:"favorite"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.SetFavorite(r.Context(), user.ID, projectID, req.Favorite); err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
