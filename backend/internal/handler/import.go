package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"quiz-platform/internal/domain"
	"quiz-platform/internal/middleware"
	"quiz-platform/internal/service/importer"
)

type ImportHandler struct {
	gf *importer.GoogleFormsImporter
}

func NewImportHandler(client *http.Client) *ImportHandler {
	return &ImportHandler{
		gf: importer.NewGoogleFormsImporter(client),
	}
}

// POST /api/import/google-forms
//
// Принимает ссылку на Google Form, возвращает структуру теста для превью.
// Тест НЕ сохраняется автоматически — клиент показывает превью,
// затем отправляет POST /api/tests для публикации.
//
// Body:   {"url": "https://docs.google.com/forms/d/.../viewform"}
// Response: ImportedTest (title, description, questions[])
func (h *ImportHandler) ImportGoogleForm(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	var req domain.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	slog.Info("import google form", "user_id", claims.UserID, "url", req.URL)

	imported, err := h.gf.Import(req.URL)
	if err != nil {
		// Различаем ошибки валидации URL от сетевых/парсинговых
		msg := err.Error()
		switch {
		case strings.Contains(msg, "not a Google Forms URL"),
			strings.Contains(msg, "empty URL"),
			strings.Contains(msg, "invalid URL"),
			strings.Contains(msg, "could not extract form ID"):
			writeError(w, http.StatusBadRequest, msg)
		case strings.Contains(msg, "FB_PUBLIC_LOAD_DATA_ not found"):
			writeError(w, http.StatusUnprocessableEntity,
				"could not parse form; make sure the form is public and the URL is correct")
		default:
			slog.Error("import google form", "user_id", claims.UserID, "url", req.URL, "err", err)
			writeError(w, http.StatusBadGateway, "failed to fetch or parse the Google Form")
		}
		return
	}

	writeJSON(w, http.StatusOK, imported)
}
