package handler

import (
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

// POST /api/admin/tests/:id/official
func (h *AdminHandler) MakeOfficial(w http.ResponseWriter, r *http.Request) {
	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}

	if err := h.tests.SetOfficial(r.Context(), testID, true); err != nil {
		switch err {
		case domain.ErrTestNotFound:
			writeError(w, http.StatusNotFound, "test not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update test")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
