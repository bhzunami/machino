package handler

import (
	"net/http"
	"time"

	"machino/internal/model"

	"github.com/gorilla/mux"
)

func (h *Handler) listTodos(w http.ResponseWriter, r *http.Request, _ model.User) {
	projectID := mux.Vars(r)["projectID"]
	todos, err := h.store.ListTodos(r.Context(), projectID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"todos": todos})
}

func (h *Handler) createTodo(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		DueDate     *string `json:"dueDate"`
		Priority    string  `json:"priority"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	dueDate, ok := parseOptionalDate(w, req.DueDate)
	if !ok {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	todo, err := h.store.CreateTodo(r.Context(), user.ID, projectID, req.Title, req.Description, req.Priority, dueDate)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusCreated, map[string]any{"todo": todo})
}

func (h *Handler) updateTodo(w http.ResponseWriter, r *http.Request, _ model.User) {
	var req struct {
		Completed   *bool   `json:"completed"`
		Title       *string `json:"title"`
		Description *string `json:"description"`
		DueDate     *string `json:"dueDate"`
		Priority    *string `json:"priority"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	var duePtr **time.Time
	if req.DueDate != nil {
		due, ok := parseOptionalDate(w, req.DueDate)
		if !ok {
			return
		}
		duePtr = &due
	}
	todo, err := h.store.UpdateTodo(r.Context(), mux.Vars(r)["todoID"], req.Completed, req.Title, req.Description, req.Priority, duePtr)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(todo.ProjectID)
	respond(w, http.StatusOK, map[string]any{"todo": todo})
}

func (h *Handler) deleteCompletedTodos(w http.ResponseWriter, r *http.Request, _ model.User) {
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.DeleteCompletedTodos(r.Context(), projectID); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) reorderTodos(w http.ResponseWriter, r *http.Request, _ model.User) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.ReorderTodos(r.Context(), projectID, req.IDs); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) projectWS(w http.ResponseWriter, r *http.Request, _ model.User) {
	h.hub.Serve(w, r, mux.Vars(r)["projectID"])
}
