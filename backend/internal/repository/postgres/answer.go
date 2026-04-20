package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quiz-platform/internal/domain"
)

type AnswerRepository struct {
	db *pgxpool.Pool
}

func NewAnswerRepository(db *pgxpool.Pool) *AnswerRepository {
	return &AnswerRepository{db: db}
}

// GetUserHistory возвращает историю ответов пользователя с пагинацией.
func (r *AnswerRepository) GetUserHistory(ctx context.Context, userID int64, limit, offset int) (*domain.PaginatedAnswers, error) {
	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM test_answers WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count answers: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT ta.id, ta.test_id, t.title, ta.answers, ta.result, ta.score, ta.created_at, ta.updated_at
		 FROM test_answers ta
		 JOIN tests t ON t.id = ta.test_id
		 WHERE ta.user_id = $1
		 ORDER BY ta.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	items := make([]domain.AnswerHistoryItem, 0)
	for rows.Next() {
		var item domain.AnswerHistoryItem
		if err := rows.Scan(
			&item.AnswerID, &item.TestID, &item.TestTitle,
			&item.Answers, &item.Result, &item.Score,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan answer row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &domain.PaginatedAnswers{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// GetUserAnswer возвращает ответы пользователя на конкретный тест.
// Возвращает nil, nil если ответа ещё нет — хэндлер отдаст 200 с null.
func (r *AnswerRepository) GetUserAnswer(ctx context.Context, userID, testID int64) (*domain.AnswerHistoryItem, error) {
	item := &domain.AnswerHistoryItem{}
	err := r.db.QueryRow(ctx,
		`SELECT ta.id, ta.test_id, t.title, ta.answers, ta.result, ta.score, ta.created_at, ta.updated_at
		 FROM test_answers ta
		 JOIN tests t ON t.id = ta.test_id
		 WHERE ta.user_id = $1 AND ta.test_id = $2`,
		userID, testID,
	).Scan(
		&item.AnswerID, &item.TestID, &item.TestTitle,
		&item.Answers, &item.Result, &item.Score,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user answer: %w", err)
	}
	return item, nil
}
