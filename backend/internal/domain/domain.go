package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// -------------------------------------------------------
// Ошибки
// -------------------------------------------------------

var (
	ErrTestNotFound    = errors.New("test not found")
	ErrAlreadyAnswered = errors.New("already answered") // оставлен для совместимости
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("not found")
)

// -------------------------------------------------------
// Вопросы
// -------------------------------------------------------

// QuestionType — типы вопросов платформы.
type QuestionType string

const (
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeScale          QuestionType = "scale"
	QuestionTypeText           QuestionType = "text"
	QuestionTypeVectorScale    QuestionType = "vector_scale"
)

// QuestionOption — вариант ответа для single/multiple choice.
type QuestionOption struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Score *int   `json:"score,omitempty"`
}

// Question — один вопрос теста.
type Question struct {
	ID       string           `json:"id"`
	Text     string           `json:"text"`
	Type     QuestionType     `json:"type"`
	Required bool             `json:"required"`
	Options  []QuestionOption `json:"options,omitempty"`
	MinValue *int             `json:"min,omitempty"`
	MaxValue *int             `json:"max,omitempty"`
	MinLabel string           `json:"min_label,omitempty"`
	MaxLabel string           `json:"max_label,omitempty"`
	Rows     []string         `json:"rows,omitempty"`
	Cols     []string         `json:"cols,omitempty"`
}

// -------------------------------------------------------
// Тесты
// -------------------------------------------------------

// Test — тест из БД, возвращаемый клиенту.
type Test struct {
	ID           int64           `json:"id"`
	AuthorID     int64           `json:"author_id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Questions    json.RawMessage `json:"questions"`
	Tags         []string        `json:"tags"`
	ResultConfig json.RawMessage `json:"result_config,omitempty"`
	Status       string          `json:"status"`
	IsOfficial   bool            `json:"is_official"`
	Rating       int             `json:"rating"`
	PassCount    int             `json:"pass_count"`
	CommentCount int             `json:"comment_count"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// TestListParams — параметры для листинга тестов.
type TestListParams struct {
	Sort     string // "rating" | "newest" | "pass_count" | "comments"
	Filter   string // "all" | "official" | "my"
	Search   string
	Tags     []string
	AuthorID int64
	UserID   int64 // для фильтра "my"
	Limit    int
	Offset   int
}

// TestListResponse — ответ со списком тестов и пагинацией.
type TestListResponse struct {
	Items  []Test `json:"items"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// CreateTestRequest — тело запроса на создание теста.
type CreateTestRequest struct {
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Questions    json.RawMessage `json:"questions"`
	Tags         []string        `json:"tags"`
	ResultConfig json.RawMessage `json:"result_config,omitempty"`
}

// SubmitAnswerRequest — тело запроса на отправку ответов.
type SubmitAnswerRequest struct {
	Answers map[string]interface{} `json:"answers"`
}

// AnswerResult — результат прохождения теста.
type AnswerResult struct {
	AnswerID  int64           `json:"answer_id"`
	TestID    int64           `json:"test_id"`
	Score     *int            `json:"score,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"` // v1: null, v2: вычисленный
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Comment — комментарий к тесту.
type Comment struct {
	ID        int64     `json:"id"`
	TestID    int64     `json:"test_id"`
	UserID    int64     `json:"user_id"`
	Nickname  string    `json:"nickname"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// -------------------------------------------------------
// Профиль и история
// -------------------------------------------------------

// Profile — демографический профиль пользователя.
type Profile struct {
	UserID    int64     `json:"user_id"`
	Age       *int      `json:"age,omitempty"`
	Gender    *string   `json:"gender,omitempty"` // "M" | "F" | "O"
	Income    *int      `json:"income,omitempty"`
	Children  *int      `json:"children,omitempty"`
	Religion  *string   `json:"religion,omitempty"`
	Education *string   `json:"education,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateProfileRequest — поля для обновления профиля.
type UpdateProfileRequest struct {
	Age       *int    `json:"age"`
	Gender    *string `json:"gender"`
	Income    *int    `json:"income"`
	Children  *int    `json:"children"`
	Religion  *string `json:"religion"`
	Education *string `json:"education"`
}

// AnswerHistoryItem — одна запись в истории ответов пользователя.
type AnswerHistoryItem struct {
	AnswerID  int64           `json:"answer_id"`
	TestID    int64           `json:"test_id"`
	TestTitle string          `json:"test_title"`
	Answers   json.RawMessage `json:"answers"`
	Result    json.RawMessage `json:"result,omitempty"` // v1: null, v2: вычисленный
	Score     *int            `json:"score,omitempty"`  // legacy
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// PaginatedAnswers — ответ со списком и метаданными пагинации.
type PaginatedAnswers struct {
	Items  []AnswerHistoryItem `json:"items"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

// -------------------------------------------------------
// Импорт тестов
// -------------------------------------------------------

// ImportedTest — результат импорта, возвращается клиенту для превью.
type ImportedTest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Questions   []Question `json:"questions"`
	SourceURL   string     `json:"source_url"`
	SourceType  string     `json:"source_type"`
}

// ImportRequest — тело запроса на импорт.
type ImportRequest struct {
	URL string `json:"url"`
}

// -------------------------------------------------------
// Auth
// -------------------------------------------------------

// AuthClaims — данные пользователя из JWT-токена.
type AuthClaims struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	IsAdmin  bool   `json:"is_admin"`
}

// -------------------------------------------------------
// Интерфейсы репозиториев
// -------------------------------------------------------

type TestRepository interface {
	List(ctx context.Context, params TestListParams) (*TestListResponse, error)
	GetByID(ctx context.Context, id int64) (*Test, error)
	Create(ctx context.Context, authorID int64, req CreateTestRequest) (*Test, error)
	SubmitAnswer(ctx context.Context, testID, userID int64, req SubmitAnswerRequest, result json.RawMessage) (*AnswerResult, error)
	Vote(ctx context.Context, testID, userID int64, vote int) error
	GetComments(ctx context.Context, testID int64) ([]Comment, error)
	AddComment(ctx context.Context, testID, userID int64, nickname, content string) (*Comment, error)
	DeleteComment(ctx context.Context, testID, commentID, userID int64) error
	SetOfficial(ctx context.Context, testID int64, official bool) error
}

type AnswerRepository interface {
	GetUserHistory(ctx context.Context, userID int64, limit, offset int) (*PaginatedAnswers, error)
	GetUserAnswer(ctx context.Context, userID, testID int64) (*AnswerHistoryItem, error)
}
