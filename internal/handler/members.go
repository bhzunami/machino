package handler

import (
	"net/http"

	"github.com/bhzunami/machino/internal/model"

	"github.com/gorilla/mux"
)

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	members, err := h.store.ListMembers(r.Context(), user.ID, projectID)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusOK, members)
}

func (h *Handler) addMember(w http.ResponseWriter, r *http.Request, user model.User) {
	projectID := mux.Vars(r)["projectID"]
	var body struct {
		Email  string `json:"email"`
		UserID string `json:"userId"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	var (
		member model.ProjectMember
		err    error
	)
	if body.UserID != "" {
		member, err = h.store.AddMemberByUserID(r.Context(), user.ID, projectID, body.UserID)
	} else {
		member, err = h.store.AddMember(r.Context(), user.ID, projectID, body.Email)
	}
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	respond(w, http.StatusCreated, member)
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request, user model.User) {
	vars := mux.Vars(r)
	projectID := vars["projectID"]
	memberUserID := vars["memberUserID"]
	if err := h.store.RemoveMember(r.Context(), user.ID, projectID, memberUserID); err != nil {
		h.handleStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
