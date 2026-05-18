package alert

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
	mux.HandleFunc("POST /api/products/{product_id}/alerts", h.create)
	mux.HandleFunc("GET /api/products/{product_id}/alerts", h.list)
	mux.HandleFunc("PATCH /api/products/{product_id}/alerts/{alert_id}", h.update)
	mux.HandleFunc("DELETE /api/products/{product_id}/alerts/{alert_id}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body alertRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	alert, err := h.store.Create(r.Context(), r.PathValue("product_id"), CreateInput{
		Name:               body.NameString(),
		ConditionType:      body.ConditionTypeString(),
		TargetUnit:         body.TargetUnitString(),
		ThresholdValueText: body.ThresholdValueTextString(),
		IsActive:           body.IsActive,
	})
	if err != nil {
		handleAlertError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, alert)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.store.List(r.Context(), r.PathValue("product_id"))
	if err != nil {
		handleAlertError(w, err)
		return
	}
	if alerts == nil {
		alerts = []Alert{}
	}
	writeJSON(w, http.StatusOK, map[string][]Alert{"alerts": alerts})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body alertRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	alert, err := h.store.Update(r.Context(), r.PathValue("product_id"), r.PathValue("alert_id"), UpdateInput{
		Name:               body.Name,
		ConditionType:      body.ConditionType,
		TargetUnit:         body.TargetUnit,
		ThresholdValueText: body.ThresholdValueText,
		IsActive:           body.IsActive,
	})
	if err != nil {
		handleAlertError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, alert)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("product_id"), r.PathValue("alert_id")); err != nil {
		handleAlertError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type alertRequest struct {
	Name               *string `json:"name"`
	ConditionType      *string `json:"condition_type"`
	TargetUnit         *string `json:"target_unit"`
	ThresholdValueText *string `json:"threshold_value_text"`
	IsActive           *bool   `json:"is_active"`
}

func (r alertRequest) NameString() string {
	if r.Name == nil {
		return ""
	}
	return *r.Name
}

func (r alertRequest) ConditionTypeString() string {
	if r.ConditionType == nil {
		return ""
	}
	return *r.ConditionType
}

func (r alertRequest) TargetUnitString() string {
	if r.TargetUnit == nil {
		return ""
	}
	return *r.TargetUnit
}

func (r alertRequest) ThresholdValueTextString() string {
	if r.ThresholdValueText == nil {
		return ""
	}
	return *r.ThresholdValueText
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

func handleAlertError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeErrorWithField(w, http.StatusBadRequest, "invalid_input", err.Error(), validationErr.Field)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "alert resource not found")
	default:
		slog.Default().Error("alert request failed", slog.String("error", err.Error()))
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
