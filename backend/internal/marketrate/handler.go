package marketrate

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
	mux.HandleFunc("POST /api/market-rates", h.create)
	mux.HandleFunc("GET /api/market-rates/latest", h.latest)
	mux.HandleFunc("GET /api/market-rates/history", h.history)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body marketRateRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	input, err := body.createInput()
	if err != nil {
		handleMarketRateError(w, err)
		return
	}

	rate, err := h.store.Create(r.Context(), input)
	if err != nil {
		handleMarketRateError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, rate)
}

func (h *Handler) latest(w http.ResponseWriter, r *http.Request) {
	rates, err := h.store.Latest(r.Context(), optionalRateType(r))
	if err != nil {
		handleMarketRateError(w, err)
		return
	}
	if rates == nil {
		rates = []MarketRate{}
	}
	writeJSON(w, http.StatusOK, map[string][]MarketRate{"market_rates": rates})
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	rates, err := h.store.History(r.Context(), optionalRateType(r))
	if err != nil {
		handleMarketRateError(w, err)
		return
	}
	if rates == nil {
		rates = []MarketRate{}
	}
	writeJSON(w, http.StatusOK, map[string][]MarketRate{"market_rates": rates})
}

type marketRateRequest struct {
	RateType   *string `json:"rate_type"`
	ValueText  *string `json:"value_text"`
	ObservedAt *string `json:"observed_at"`
}

func (r marketRateRequest) createInput() (CreateInput, error) {
	rateType := ""
	if r.RateType != nil {
		rateType = *r.RateType
	}
	valueText := ""
	if r.ValueText != nil {
		valueText = *r.ValueText
	}

	var observedAt time.Time
	if r.ObservedAt != nil {
		parsed, err := time.Parse(time.RFC3339Nano, *r.ObservedAt)
		if err != nil {
			return CreateInput{}, fieldError("observed_at", "must be an RFC3339 timestamp")
		}
		observedAt = parsed
	}

	return NormalizeCreate(CreateInput{
		RateType:   rateType,
		ValueText:  valueText,
		ObservedAt: observedAt,
	})
}

func optionalRateType(r *http.Request) *string {
	value := r.URL.Query().Get("rate_type")
	if value == "" {
		return nil
	}
	return &value
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

func handleMarketRateError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeErrorWithField(w, http.StatusBadRequest, "invalid_input", err.Error(), validationErr.Field)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "market rate not found")
	default:
		slog.Default().Error("market rate request failed", slog.String("error", err.Error()))
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
