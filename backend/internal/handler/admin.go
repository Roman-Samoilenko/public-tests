package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"quiz-platform/internal/domain"
)

type AdminHandler struct {
	tests domain.TestRepository
}

func NewAdminHandler(tests domain.TestRepository) *AdminHandler {
	return &AdminHandler{tests: tests}
}

// SetOfficial PATCH /api/admin/tests/:id/official
// Тело запроса: { "official": true } или { "official": false }
func (h *AdminHandler) SetOfficial(w http.ResponseWriter, r *http.Request) {
	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}

	var body struct {
		Official bool `json:"official"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.tests.SetOfficial(r.Context(), testID, body.Official); err != nil {
		switch {
		case errors.Is(err, domain.ErrTestNotFound):
			writeError(w, http.StatusNotFound, "test not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update test")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// SetStatus PATCH /api/admin/tests/:id/status
// Тело запроса: { "status": "published" } или { "status": "blocked" }
func (h *AdminHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Допустимые значения статуса
	if body.Status != "published" && body.Status != "blocked" {
		writeError(w, http.StatusBadRequest, "status must be 'published' or 'blocked'")
		return
	}

	if err := h.tests.SetStatus(r.Context(), testID, body.Status); err != nil {
		switch {
		case errors.Is(err, domain.ErrTestNotFound):
			writeError(w, http.StatusNotFound, "test not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update test")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
