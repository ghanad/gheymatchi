package source

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/products/{product_id}/sources", h.create)
	mux.HandleFunc("GET /api/products/{product_id}/sources", h.list)
	mux.HandleFunc("PATCH /api/products/{product_id}/sources/{source_id}", h.update)
	mux.HandleFunc("DELETE /api/products/{product_id}/sources/{source_id}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body sourceRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	productSource, err := h.store.Create(r.Context(), r.PathValue("product_id"), CreateInput{
		URL:        body.URLString(),
		SourceName: body.SourceNameString(),
		IsActive:   body.IsActive,
	})
	if err != nil {
		handleSourceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, productSource)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	sources, err := h.store.List(r.Context(), r.PathValue("product_id"))
	if err != nil {
		handleSourceError(w, err)
		return
	}
	if sources == nil {
		sources = []ProductSource{}
	}
	writeJSON(w, http.StatusOK, map[string][]ProductSource{"sources": sources})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body sourceRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	productSource, err := h.store.Update(r.Context(), r.PathValue("product_id"), r.PathValue("source_id"), UpdateInput{
		URL:        body.URL,
		SourceName: body.SourceName,
		IsActive:   body.IsActive,
	})
	if err != nil {
		handleSourceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, productSource)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("product_id"), r.PathValue("source_id")); err != nil {
		handleSourceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type sourceRequest struct {
	URL        *string `json:"url"`
	SourceName *string `json:"source_name"`
	IsActive   *bool   `json:"is_active"`
}

func (r sourceRequest) URLString() string {
	if r.URL == nil {
		return ""
	}
	return *r.URL
}

func (r sourceRequest) SourceNameString() string {
	if r.SourceName == nil {
		return ""
	}
	return *r.SourceName
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

func handleSourceError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeErrorWithField(w, http.StatusBadRequest, "invalid_input", err.Error(), validationErr.Field)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "product source not found")
	default:
		slog.Default().Error("product source request failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeErrorWithField(w, status, code, message, "")
}

func writeErrorWithField(w http.ResponseWriter, status int, code string, message string, field string) {
	response := errorResponse{
		Error: responseError{
			Code:    code,
			Message: message,
			Field:   field,
		},
	}
	writeJSON(w, status, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Error("write response", slog.String("error", err.Error()))
	}
}
