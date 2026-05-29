package handler

import (
	"net/http"

	"github.com/bhzunami/machino/internal/model"

	"github.com/gorilla/mux"
)

func (h *Handler) listColumns(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	cols, err := h.store.ListColumns(r.Context(), user.ID, projectID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"columns": cols})
}

func (h *Handler) createColumn(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Title string `json:"title"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	col, err := h.store.CreateColumn(r.Context(), user.ID, projectID, req.Title)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusCreated, map[string]any{"column": col})
}

func (h *Handler) updateColumn(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Title string `json:"title"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	col, err := h.store.UpdateColumn(r.Context(), user.ID, mux.Vars(r)["columnID"], req.Title)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(col.ProjectID)
	respond(w, http.StatusOK, map[string]any{"column": col})
}

func (h *Handler) deleteColumn(w http.ResponseWriter, r *http.Request, user model.User) {
	columnID := mux.Vars(r)["columnID"]
	projectID, err := h.store.DeleteColumn(r.Context(), user.ID, columnID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) reorderColumns(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.ReorderColumns(r.Context(), user.ID, projectID, req.IDs); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
