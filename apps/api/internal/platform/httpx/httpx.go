// Package httpx owns JSON transport conventions and domain-error serialization.
package httpx

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/lahiruudayakumara/rail-seat-booking/apps/api/internal/platform/apperror"
)

type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"requestId"`
}
type Page[T any] struct {
	Items []T      `json:"items"`
	Page  PageInfo `json:"page"`
}
type PageInfo struct {
	NextCursor *string `json:"nextCursor,omitempty"`
	HasMore    bool    `json:"hasMore"`
}

func PageOf[T any](items []T) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Page: PageInfo{}}
}
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func WriteError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		appErr = apperror.Wrap(err)
	}
	if appErr.Status >= 500 {
		logger.Error("request failed", "error", err, "request_id", RequestID(r))
	}
	WriteJSON(w, appErr.Status, ErrorResponse{appErr.Code, appErr.Message, appErr.Details, RequestID(r)})
}
func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return apperror.Validation("body", "Invalid JSON body.")
	}
	return nil
}
func PathUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperror.Validation(name, "Must be a UUID.")
	}
	return id, nil
}
func QueryUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(r.URL.Query().Get(name))
	if err != nil {
		return uuid.Nil, apperror.Validation(name, "Must be a UUID.")
	}
	return id, nil
}
func RequestID(r *http.Request) string {
	raw := chimiddleware.GetReqID(r.Context())
	if id, err := uuid.Parse(raw); err == nil {
		return id.String()
	}
	sum := sha256.Sum256([]byte(raw))
	id, _ := uuid.FromBytes(sum[:16])
	return id.String()
}
