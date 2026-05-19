package notification

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"gheymatchi/backend/internal/auth"
)

type Handler struct {
	store UserStore
}

func NewHandler(store UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/notifications", h.list)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.RequireUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication required")
		return
	}
	notifications, err := h.store.ListForUser(r.Context(), userID)
	if err != nil {
		slog.Default().Error("notification request failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	if notifications == nil {
		notifications = []Notification{}
	}
	writeJSON(w, http.StatusOK, map[string][]Notification{"notifications": notifications})
}

type errorResponse struct {
	Error responseError `json:"error"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{Error: responseError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("write response", slog.String("error", err.Error()))
	}
}
