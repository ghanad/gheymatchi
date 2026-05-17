package product

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
	mux.HandleFunc("POST /api/products", h.create)
	mux.HandleFunc("GET /api/products", h.list)
	mux.HandleFunc("GET /api/products/{id}", h.get)
	mux.HandleFunc("PATCH /api/products/{id}", h.update)
	mux.HandleFunc("DELETE /api/products/{id}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body productRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	product, err := h.store.Create(r.Context(), CreateInput{
		Name:        body.NameString(),
		Description: body.DescriptionValue(),
	})
	if err != nil {
		handleProductError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.List(r.Context())
	if err != nil {
		handleProductError(w, err)
		return
	}
	if products == nil {
		products = []Product{}
	}
	writeJSON(w, http.StatusOK, map[string][]Product{"products": products})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	product, err := h.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		handleProductError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body productRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	product, err := h.store.Update(r.Context(), r.PathValue("id"), UpdateInput{
		Name:        body.NameValue(),
		Description: body.Description,
	})
	if err != nil {
		handleProductError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		handleProductError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type productRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (r productRequest) NameValue() *string {
	return r.Name
}

func (r productRequest) NameString() string {
	if r.Name == nil {
		return ""
	}
	return *r.Name
}

func (r productRequest) DescriptionValue() string {
	if r.Description == nil {
		return ""
	}
	return *r.Description
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

func handleProductError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeValidationError(w, validationErr)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "product not found")
	default:
		slog.Default().Error("product request failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeValidationError(w http.ResponseWriter, err ValidationError) {
	writeErrorWithField(w, http.StatusBadRequest, "invalid_input", err.Error(), err.Field)
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
