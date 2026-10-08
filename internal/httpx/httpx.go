package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/sahayak/sahayak/internal/domain"
)

// ErrorEnvelope is the standard error response body.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the detail of a standard error response.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Pagination describes a paginated response.
type Pagination struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

// WriteJSON writes a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// WriteError writes a standard error envelope.
func WriteError(w http.ResponseWriter, err error) {
	apiErr := domain.AsAPIError(err)
	status := apiErr.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}
	WriteJSON(w, status, ErrorEnvelope{Error: ErrorBody{
		Code:    apiErr.Code,
		Message: apiErr.Message,
		Details: apiErr.Details,
	}})
}

// DecodeJSON decodes a JSON body, rejecting unknown fields and oversized bodies.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return domain.InvalidInput("Malformed JSON request body")
	}
	return nil
}

// DecodeJSONOptional decodes a JSON body, treating an empty body as valid.
func DecodeJSONOptional(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return domain.InvalidInput("Malformed JSON request body")
	}
	return nil
}

// PageParams extracts page and per_page from the query string with defaults.
func PageParams(r *http.Request) (page, perPage, offset int) {
	page = 1
	perPage = 20
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			perPage = n
		}
	}
	offset = (page - 1) * perPage
	return page, perPage, offset
}

// IsNotFound reports whether err is a not-found API error.
func IsNotFound(err error) bool {
	var apiErr *domain.APIError
	return errors.As(err, &apiErr) && apiErr.Code == domain.CodeNotFound
}
