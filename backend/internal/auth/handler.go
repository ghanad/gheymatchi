package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/logout", h.logout)
	mux.HandleFunc("GET /api/auth/me", h.me)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body authRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	session, err := h.store.Register(r.Context(), RegisterInput{Email: body.Email, Password: body.Password})
	if err != nil {
		handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body authRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	session, err := h.store.Login(r.Context(), LoginInput{Email: body.Email, Password: body.Password})
	if err != nil {
		handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Logout(r.Context(), bearerToken(r)); err != nil {
		handleAuthError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	userID, err := RequireUserID(r.Context())
	if err != nil {
		handleAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": userID})
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func Middleware(store Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/register" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := store.AuthenticateToken(r.Context(), bearerToken(r))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithUserID(r.Context(), user.ID)))
	})
}

func decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

type errorResponse struct {
	Error responseError `json:"error"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func handleAuthError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeErrorWithField(w, http.StatusBadRequest, "invalid_input", validationErr.Error(), validationErr.Field)
	case errors.Is(err, ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
	case errors.Is(err, ErrUnauthenticated):
		writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication required")
	default:
		slog.Default().Error("auth request failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeErrorWithField(w, status, code, message, "")
}

func writeErrorWithField(w http.ResponseWriter, status int, code string, message string, field string) {
	writeJSON(w, status, errorResponse{Error: responseError{Code: code, Message: message, Field: field}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("write response", slog.String("error", err.Error()))
	}
}
