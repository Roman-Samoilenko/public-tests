package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"quiz-platform/internal/domain"
	authmw "quiz-platform/internal/middleware"
	"quiz-platform/internal/scoring"
)

type TestHandler struct {
	tests   domain.TestRepository
	answers domain.AnswerRepository
}

func NewTestHandler(tests domain.TestRepository, answers domain.AnswerRepository) *TestHandler {
	return &TestHandler{tests: tests, answers: answers}
}

// ListTests GET /api/tests?sort=rating&limit=12&offset=0&search=...&tags=...&official=1&my=1&author_id=...&status=(published|blocked)
func (h *TestHandler) ListTests(w http.ResponseWriter, r *http.Request) {
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "rating"
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 12
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	search := r.URL.Query().Get("search")

	var userID int64
	var isAdmin bool
	if claims := authmw.ClaimsFromContext(r.Context()); claims != nil {
		userID = claims.UserID
		isAdmin = claims.IsAdmin
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "published"
	} else if status != "published" {
		if !isAdmin {
			writeError(w, http.StatusForbidden, "only admins can filter by non-published status")
			return
		}
		if status != "blocked" && status != "all" {
			writeError(w, http.StatusBadRequest, "invalid status value")
			return
		}
	}

	var tags []string
	if tagsRaw := r.URL.Query().Get("tags"); tagsRaw != "" {
		tags = strings.Split(tagsRaw, ",")
	}

	authorID, _ := strconv.ParseInt(r.URL.Query().Get("author_id"), 10, 64)

	// Определяем фильтр и userID для "my"
	filter := "all"

	if r.URL.Query().Get("official") == "1" {
		filter = "official"
	} else if r.URL.Query().Get("my") == "1" {
		filter = "my"
	}

	result, err := h.tests.List(r.Context(), domain.TestListParams{
		Sort:     sort,
		Filter:   filter,
		Search:   search,
		Tags:     tags,
		AuthorID: authorID,
		UserID:   userID,
		Limit:    limit,
		Offset:   offset,
		Status:   status,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch tests")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetTest GET /api/tests/:id.
func (h *TestHandler) GetTest(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}
	test, err := h.tests.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrTestNotFound) {
		writeError(w, http.StatusNotFound, "test not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch test")
		return
	}
	writeJSON(w, http.StatusOK, test)
}

// CreateTest POST /api/tests.
func (h *TestHandler) CreateTest(w http.ResponseWriter, r *http.Request) {
	userID := authmw.ClaimsFromContext(r.Context()).UserID

	var body domain.CreateTestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(body.Questions) == 0 {
		writeError(w, http.StatusBadRequest, "questions are required")
		return
	}

	test, err := h.tests.Create(r.Context(), userID, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create test")
		return
	}
	writeJSON(w, http.StatusCreated, test)
}

// GetMyAnswer GET /api/tests/:id/my-answer.
func (h *TestHandler) GetMyAnswer(w http.ResponseWriter, r *http.Request) {
	userID := authmw.ClaimsFromContext(r.Context()).UserID

	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}

	answer, err := h.answers.GetUserAnswer(r.Context(), userID, testID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch answer")
		return
	}
	// answer == nil значит ответа нет — отдаём 200 с null
	writeJSON(w, http.StatusOK, answer)
}

// SubmitAnswer POST /api/tests/:id/answers.
func (h *TestHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	userID := authmw.ClaimsFromContext(r.Context()).UserID

	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}

	var body domain.SubmitAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 1. Загружаем тест для получения Questions и ResultConfig
	test, err := h.tests.GetByID(r.Context(), testID)
	if err != nil {
		if errors.Is(err, domain.ErrTestNotFound) {
			writeError(w, http.StatusNotFound, "test not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch test")
		}
		return
	}

	// 2. Вызываем ядро подсчёта из пакета scoring
	scoringResult, err := scoring.Calculate(
		test.ResultConfig, // json.RawMessage
		test.Questions,    // json.RawMessage
		body.Answers,      // map[string]interface{}
	)
	if err != nil {
		// ошибка парсинга конфигурации или вопросов — внутренняя ошибка
		writeError(w, http.StatusInternalServerError, "scoring failed")
		return
	}

	// 3. Сохраняем ответ вместе с результатом (может быть nil)
	ar, err := h.tests.SubmitAnswer(r.Context(), testID, userID, body, scoringResult)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTestNotFound):
			writeError(w, http.StatusNotFound, "test not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to submit answer")
		}
		return
	}

	writeJSON(w, http.StatusCreated, ar)
}

// VoteTest POST /api/tests/:id/vote.
func (h *TestHandler) VoteTest(w http.ResponseWriter, r *http.Request) {
	userID := authmw.ClaimsFromContext(r.Context()).UserID

	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}
	var body struct {
		Vote int `json:"vote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Vote != 1 && body.Vote != -1 {
		writeError(w, http.StatusBadRequest, "vote must be 1 or -1")
		return
	}

	if err := h.tests.Vote(r.Context(), testID, userID, body.Vote); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to vote")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetComments GetComments GET /api/tests/:id/comments.
func (h *TestHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}
	comments, err := h.tests.GetComments(r.Context(), testID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch comments")
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

// AddComment POST /api/tests/:id/comments.
func (h *TestHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	claims := authmw.ClaimsFromContext(r.Context())
	userID := claims.UserID

	// Никнейм берём из JWT; если пустой — fallback
	nickname := claims.Nickname
	if nickname == "" {
		nickname = "user_" + strconv.FormatInt(userID, 10)
	}

	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	comment, err := h.tests.AddComment(r.Context(), testID, userID, nickname, body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add comment")
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

// DELETE /api/tests/:id/comments/:commentId.
func (h *TestHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := authmw.ClaimsFromContext(r.Context()).UserID

	testID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}
	commentID, err := parseID(chi.URLParam(r, "commentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	if err := h.tests.DeleteComment(r.Context(), testID, commentID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "comment not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete comment")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateTest PATCH /api/tests/:id
func (h *TestHandler) UpdateTest(w http.ResponseWriter, r *http.Request) {
	claims := authmw.ClaimsFromContext(r.Context())
	userID := claims.UserID
	isAdmin := claims.IsAdmin

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid test id")
		return
	}

	var body domain.CreateTestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(body.Questions) == 0 || string(body.Questions) == "null" || string(body.Questions) == "[]" {
		writeError(w, http.StatusBadRequest, "questions are required")
		return
	}

	test, err := h.tests.UpdateTest(r.Context(), id, userID, isAdmin, body)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTestNotFound):
			writeError(w, http.StatusNotFound, "test not found")
		case errors.Is(err, domain.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update test")
		}
		return
	}
	writeJSON(w, http.StatusOK, test)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
