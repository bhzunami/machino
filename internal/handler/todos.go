package handler

import (
	"net/http"
	"time"

	"github.com/bhzunami/machino/internal/model"
	"github.com/bhzunami/machino/internal/store"

	"github.com/gorilla/mux"
)

func (h *Handler) listTodos(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	todos, err := h.store.ListTodos(r.Context(), user.ID, projectID)
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
		ColumnID    *string `json:"columnId"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	dueDate, ok := parseOptionalDate(w, req.DueDate)
	if !ok {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	todo, err := h.store.CreateTodo(r.Context(), user.ID, projectID, req.Title, req.Description, req.Priority, dueDate, req.ColumnID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusCreated, map[string]any{"todo": todo})
}

func (h *Handler) updateTodo(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		Completed   *bool   `json:"completed"`
		Title       *string `json:"title"`
		Description *string `json:"description"`
		DueDate     *string `json:"dueDate"`
		Priority    *string `json:"priority"`
		ColumnID    *string `json:"columnId"`
		ClearColumn bool    `json:"clearColumn"`
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
	var colUpdate *store.ColumnUpdate
	if req.ColumnID != nil || req.ClearColumn {
		colUpdate = &store.ColumnUpdate{Set: true, Val: req.ColumnID}
	}
	todo, err := h.store.UpdateTodo(r.Context(), user.ID, mux.Vars(r)["todoID"], req.Completed, req.Title, req.Description, req.Priority, duePtr, colUpdate)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(todo.ProjectID)
	respond(w, http.StatusOK, map[string]any{"todo": todo})
}

func (h *Handler) deleteTodo(w http.ResponseWriter, r *http.Request, user model.User) {
	todoID := mux.Vars(r)["todoID"]
	projectID, err := h.store.DeleteTodo(r.Context(), user.ID, todoID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) deleteCompletedTodos(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.DeleteCompletedTodos(r.Context(), user.ID, projectID); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) reorderTodos(w http.ResponseWriter, r *http.Request, user model.User) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	projectID := mux.Vars(r)["projectID"]
	if err := h.store.ReorderTodos(r.Context(), user.ID, projectID, req.IDs); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Broadcast(projectID)
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) projectWS(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	// Verify the user owns this project before allowing WebSocket access.
	if _, err := h.store.GetProject(r.Context(), user.ID, projectID); err != nil {
		h.handleStoreError(w, err)
		return
	}
	h.hub.Serve(w, r, projectID)
}
