package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"quiz-platform/internal/domain"
)

type TestRepository struct {
	db *pgxpool.Pool
}

func NewTestRepository(db *pgxpool.Pool) *TestRepository {
	return &TestRepository{db: db}
}

// queryBuilder — вспомогательный тип для безопасного построения динамических WHERE.
type queryBuilder struct {
	conds  []string
	args   []any
	argIdx int
}

func (b *queryBuilder) add(cond string, val any) {
	b.argIdx++
	b.conds = append(b.conds, fmt.Sprintf(cond, b.argIdx))
	b.args = append(b.args, val)
}

func (b *queryBuilder) where() string {
	if len(b.conds) == 0 {
		return ""
	}
	return "AND " + strings.Join(b.conds, " AND ")
}

// List возвращает список тестов с фильтрацией, поиском, сортировкой и пагинацией.
func (r *TestRepository) List(ctx context.Context, params domain.TestListParams) (*domain.TestListResponse, error) {
	// ORDER BY
	orderBy := "t.rating DESC"
	switch params.Sort {
	case "newest":
		orderBy = "t.created_at DESC"
	case "pass_count":
		orderBy = "t.pass_count DESC"
	case "comments":
		orderBy = "comment_count DESC"
	}

	// Динамические условия
	qb := &queryBuilder{argIdx: 0}

	if params.Search != "" {
		qb.add("t.search_vector @@ plainto_tsquery('russian', $%d)", params.Search)
	}
	if len(params.Tags) > 0 {
		qb.add("t.tags && $%d", params.Tags)
	}
	if params.Filter == "official" {
		qb.conds = append(qb.conds, "t.is_official = TRUE")
	}
	if params.Filter == "my" && params.UserID != 0 {
		qb.add("t.author_id = $%d", params.UserID)
	}
	if params.AuthorID != 0 {
		qb.add("t.author_id = $%d", params.AuthorID)
	}

	whereClause := qb.where()

	// COUNT для пагинации
	countQuery := fmt.Sprintf(
		`SELECT COUNT(*) FROM tests t WHERE t.status = 'published' %s`,
		whereClause,
	)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, qb.args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count tests: %w", err)
	}

	// Добавляем LIMIT и OFFSET после динамических аргументов
	limitIdx := qb.argIdx + 1
	offsetIdx := qb.argIdx + 2
	listArgs := append(qb.args, params.Limit, params.Offset)

	selectQuery := fmt.Sprintf(`
		SELECT
			t.id, t.author_id, t.title, t.description, t.questions,
			t.tags, t.result_config,
			t.status, t.is_official, t.rating, t.pass_count,
			t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM test_comments WHERE test_id = t.id)::int AS comment_count
		FROM tests t
		WHERE t.status = 'published'
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, limitIdx, offsetIdx,
	)

	rows, err := r.db.Query(ctx, selectQuery, listArgs...)
	if err != nil {
		return nil, fmt.Errorf("list tests: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Test, 0, params.Limit)
	for rows.Next() {
		var t domain.Test
		if err := rows.Scan(
			&t.ID, &t.AuthorID, &t.Title, &t.Description, &t.Questions,
			&t.Tags, &t.ResultConfig,
			&t.Status, &t.IsOfficial, &t.Rating, &t.PassCount,
			&t.CreatedAt, &t.UpdatedAt,
			&t.CommentCount,
		); err != nil {
			return nil, fmt.Errorf("scan test: %w", err)
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &domain.TestListResponse{
		Items:  items,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// GetByID возвращает тест по ID.
func (r *TestRepository) GetByID(ctx context.Context, id int64) (*domain.Test, error) {
	var t domain.Test
	err := r.db.QueryRow(ctx,
		`SELECT
			t.id, t.author_id, t.title, t.description, t.questions,
			t.tags, t.result_config,
			t.status, t.is_official, t.rating, t.pass_count,
			t.created_at, t.updated_at,
			(SELECT COUNT(*) FROM test_comments WHERE test_id = t.id)::int AS comment_count
		 FROM tests t
		 WHERE t.id = $1`, id,
	).Scan(
		&t.ID, &t.AuthorID, &t.Title, &t.Description, &t.Questions,
		&t.Tags, &t.ResultConfig,
		&t.Status, &t.IsOfficial, &t.Rating, &t.PassCount,
		&t.CreatedAt, &t.UpdatedAt,
		&t.CommentCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get test: %w", err)
	}
	return &t, nil
}

// Create создаёт новый тест и добавляет запись в moderation_log.
func (r *TestRepository) Create(ctx context.Context, authorID int64, req domain.CreateTestRequest) (*domain.Test, error) {
	// questions уже json.RawMessage — маршалить не нужно
	// tags: nil → пустой массив
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var t domain.Test
	err = tx.QueryRow(ctx,
		`INSERT INTO tests (author_id, title, description, questions, tags, result_config)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING
		 	id, author_id, title, description, questions,
		 	tags, result_config,
		 	status, is_official, rating, pass_count, created_at, updated_at`,
		authorID, req.Title, req.Description, req.Questions, tags, req.ResultConfig,
	).Scan(
		&t.ID, &t.AuthorID, &t.Title, &t.Description, &t.Questions,
		&t.Tags, &t.ResultConfig,
		&t.Status, &t.IsOfficial, &t.Rating, &t.PassCount,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert test: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO moderation_log (test_id, author_id, title, action)
		 VALUES ($1, $2, $3, 'created')`,
		t.ID, authorID, req.Title,
	)
	if err != nil {
		return nil, fmt.Errorf("insert moderation log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &t, nil
}

// SubmitAnswer сохраняет ответы через upsert.
// pass_count инкрементируется только при первом ответе.
func (r *TestRepository) SubmitAnswer(ctx context.Context, testID, userID int64, req domain.SubmitAnswerRequest, result json.RawMessage) (*domain.AnswerResult, error) {
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, fmt.Errorf("marshal answers: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Проверяем существование теста
	var testExists bool
	if err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tests WHERE id = $1)`, testID,
	).Scan(&testExists); err != nil {
		return nil, fmt.Errorf("check test exists: %w", err)
	}
	if !testExists {
		return nil, domain.ErrTestNotFound
	}

	// Проверяем, есть ли уже ответ (для pass_count)
	var existed bool
	if err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM test_answers WHERE test_id = $1 AND user_id = $2)`,
		testID, userID,
	).Scan(&existed); err != nil {
		return nil, fmt.Errorf("check answer exists: %w", err)
	}

	// Upsert
	ar := &domain.AnswerResult{TestID: testID}
	err = tx.QueryRow(ctx,
		`INSERT INTO test_answers (test_id, user_id, answers, result)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (test_id, user_id) DO UPDATE SET
		 	answers    = EXCLUDED.answers,
		 	result     = EXCLUDED.result,
		 	updated_at = NOW()
		 RETURNING id, created_at, updated_at`,
		testID, userID, answersJSON, result,
	).Scan(&ar.AnswerID, &ar.CreatedAt, &ar.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert answer: %w", err)
	}

	// Инкрементируем pass_count только для нового ответа
	if !existed {
		if _, err = tx.Exec(ctx,
			`UPDATE tests SET pass_count = pass_count + 1 WHERE id = $1`, testID,
		); err != nil {
			return nil, fmt.Errorf("update pass_count: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return ar, nil
}

// Vote ставит или перезаписывает голос пользователя за тест.
// Рейтинг пересчитывается триггером в БД.
func (r *TestRepository) Vote(ctx context.Context, testID, userID int64, vote int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO test_votes (user_id, test_id, vote)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, test_id) DO UPDATE SET vote = EXCLUDED.vote`,
		userID, testID, vote,
	)
	if err != nil {
		return fmt.Errorf("upsert vote: %w", err)
	}
	return nil
}

// GetComments возвращает комментарии к тесту (новые первыми).
func (r *TestRepository) GetComments(ctx context.Context, testID int64) ([]domain.Comment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, test_id, user_id, nickname, content, created_at
		 FROM test_comments
		 WHERE test_id = $1
		 ORDER BY created_at DESC`,
		testID,
	)
	if err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.TestID, &c.UserID, &c.Nickname, &c.Content, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return comments, nil
}

// AddComment добавляет комментарий к тесту, сохраняя никнейм из JWT.
func (r *TestRepository) AddComment(ctx context.Context, testID, userID int64, nickname, content string) (*domain.Comment, error) {
	c := &domain.Comment{TestID: testID, UserID: userID, Nickname: nickname, Content: content}
	err := r.db.QueryRow(ctx,
		`INSERT INTO test_comments (test_id, user_id, nickname, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		testID, userID, nickname, content,
	).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add comment: %w", err)
	}
	return c, nil
}

// DeleteComment удаляет комментарий. Удалить может только автор.
func (r *TestRepository) DeleteComment(ctx context.Context, testID, commentID, userID int64) error {
	var ownerID int64
	err := r.db.QueryRow(ctx,
		`SELECT user_id FROM test_comments WHERE id = $1 AND test_id = $2`,
		commentID, testID,
	).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get comment owner: %w", err)
	}
	if ownerID != userID {
		return domain.ErrForbidden
	}

	if _, err = r.db.Exec(ctx,
		`DELETE FROM test_comments WHERE id = $1`, commentID,
	); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}

// SetOfficial выставляет флаг is_official для теста.
func (r *TestRepository) SetOfficial(ctx context.Context, testID int64, official bool) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE tests SET is_official = $1, updated_at = NOW() WHERE id = $2`,
		official, testID,
	)
	if err != nil {
		return fmt.Errorf("set official: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTestNotFound
	}
	return nil
}

// isUniqueViolation проверяет код ошибки PostgreSQL 23505.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
