package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quiz-platform/internal/domain"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// Get возвращает профиль пользователя.
// Если профиль ещё не создавался — возвращает пустой профиль (не ошибку).
func (r *ProfileRepository) Get(ctx context.Context, userID int64) (*domain.Profile, error) {
	p := &domain.Profile{UserID: userID}
	err := r.db.QueryRow(ctx,
		`SELECT age, gender, income, children, religion, education, updated_at
		 FROM profiles WHERE user_id = $1`,
		userID,
	).Scan(
		&p.Age, &p.Gender, &p.Income,
		&p.Children, &p.Religion, &p.Education,
		&p.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		// Профиль не заполнен — возвращаем пустой
		p.UpdatedAt = time.Now()
		return p, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

// Upsert создаёт или обновляет профиль. Обновляет только переданные поля.
func (r *ProfileRepository) Upsert(
	ctx context.Context,
	userID int64,
	req domain.UpdateProfileRequest,
) (*domain.Profile, error) {
	p := &domain.Profile{UserID: userID}
	err := r.db.QueryRow(ctx,
		`INSERT INTO profiles (user_id, age, gender, income, children, religion, education)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id) DO UPDATE SET
		     age       = COALESCE($2, profiles.age),
		     gender    = COALESCE($3, profiles.gender),
		     income    = COALESCE($4, profiles.income),
		     children  = COALESCE($5, profiles.children),
		     religion  = COALESCE($6, profiles.religion),
		     education = COALESCE($7, profiles.education),
		     updated_at = NOW()
		 RETURNING age, gender, income, children, religion, education, updated_at`,
		userID,
		req.Age, req.Gender, req.Income,
		req.Children, req.Religion, req.Education,
	).Scan(
		&p.Age, &p.Gender, &p.Income,
		&p.Children, &p.Religion, &p.Education,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert profile: %w", err)
	}
	return p, nil
}
