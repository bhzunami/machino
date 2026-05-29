package handler

import (
"encoding/json"
"log/slog"
"net/http"

"github.com/gorilla/mux"
"machino/internal/model"
)

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request, user model.User) {
projectID := mux.Vars(r)["projectID"]
members, err := h.store.ListMembers(r.Context(), user.ID, projectID)
if err != nil {
slog.Error("list members", "error", err, "project_id", projectID, "user_id", user.ID)
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
if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
respondError(w, http.StatusBadRequest, "Ungültige Anfrage.")
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
slog.Error("add member", "error", err, "project_id", projectID, "user_id", user.ID)
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
slog.Error("remove member", "error", err, "project_id", projectID, "user_id", user.ID)
h.handleStoreError(w, err)
return
}
w.WriteHeader(http.StatusNoContent)
}
