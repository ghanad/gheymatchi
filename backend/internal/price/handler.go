package price

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/products/{product_id}/sources/{source_id}/price-points", h.create)
	mux.HandleFunc("GET /api/products/{product_id}/price-points", h.listByProduct)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body pricePointRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	input, err := body.createInput()
	if err != nil {
		handlePriceError(w, err)
		return
	}

	pricePoint, err := h.store.Create(r.Context(), r.PathValue("product_id"), r.PathValue("source_id"), input)
	if err != nil {
		handlePriceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, pricePoint)
}

func (h *Handler) listByProduct(w http.ResponseWriter, r *http.Request) {
	pricePoints, err := h.store.ListByProduct(r.Context(), r.PathValue("product_id"))
	if err != nil {
		handlePriceError(w, err)
		return
	}
	if pricePoints == nil {
		pricePoints = []PricePoint{}
	}
	writeJSON(w, http.StatusOK, map[string][]PricePoint{"price_points": pricePoints})
}

type pricePointRequest struct {
	PriceIRR   *int64  `json:"price_irr"`
	CapturedAt *string `json:"captured_at"`
	RawPayload *string `json:"raw_payload"`
}

func (r pricePointRequest) createInput() (CreateInput, error) {
	if r.PriceIRR == nil {
		return CreateInput{}, fieldError("price_irr", "is required")
	}

	var capturedAt time.Time
	if r.CapturedAt != nil {
		parsed, err := time.Parse(time.RFC3339Nano, *r.CapturedAt)
		if err != nil {
			return CreateInput{}, fieldError("captured_at", "must be an RFC3339 timestamp")
		}
		capturedAt = parsed
	}

	return NormalizeCreate(CreateInput{
		PriceIRR:   *r.PriceIRR,
		CapturedAt: capturedAt,
		RawPayload: r.RawPayload,
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

func handlePriceError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeErrorWithField(w, http.StatusBadRequest, "invalid_input", err.Error(), validationErr.Field)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "price resource not found")
	default:
		slog.Default().Error("price point request failed", slog.String("error", err.Error()))
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
