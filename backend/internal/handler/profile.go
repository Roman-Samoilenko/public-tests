package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"quiz-platform/internal/domain"
	"quiz-platform/internal/middleware"
	"quiz-platform/internal/repository/postgres"
)

type ProfileHandler struct {
	profiles *postgres.ProfileRepository
	answers  *postgres.AnswerRepository
}

func NewProfileHandler(
	profiles *postgres.ProfileRepository,
	answers *postgres.AnswerRepository,
) *ProfileHandler {
	return &ProfileHandler{profiles: profiles, answers: answers}
}

// GetProfile GET /api/profile
// Возвращает профиль текущего пользователя.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	profile, err := h.profiles.Get(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("get profile", "user_id", claims.UserID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get profile")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile PUT /api/profile
// Обновляет профиль. Передавать нужно только изменяемые поля.
//
//	Body: {"age": 25, "gender": "M", "education": "higher"}
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Простая валидация
	if req.Gender != nil {
		g := *req.Gender
		if g != "M" && g != "F" && g != "O" {
			writeError(w, http.StatusBadRequest, "gender must be M, F or O")
			return
		}
	}
	if req.Age != nil && (*req.Age < 1 || *req.Age > 149) {
		writeError(w, http.StatusBadRequest, "age must be between 1 and 149")
		return
	}
	if req.Children != nil && *req.Children < 0 {
		writeError(w, http.StatusBadRequest, "children must be >= 0")
		return
	}

	profile, err := h.profiles.Upsert(r.Context(), claims.UserID, req)
	if err != nil {
		slog.Error("update profile", "user_id", claims.UserID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// GetAnswerHistory GET /api/profile/answers?limit=20&offset=0
// История всех пройденных тестов текущего пользователя.
func (h *ProfileHandler) GetAnswerHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	limit := parseIntQuery(r, "limit", 20)
	offset := parseIntQuery(r, "offset", 0)

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	history, err := h.answers.GetUserHistory(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		slog.Error("get answer history", "user_id", claims.UserID, "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	writeJSON(w, http.StatusOK, history)
}

// --- helpers ---

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
