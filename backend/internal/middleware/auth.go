package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"quiz-platform/internal/domain"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// Auth возвращает middleware, которое проверяет JWT из заголовка Authorization.
// Использует тот же JWT_SECRET, что и auth-service (shared secret).
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeUnauthorized(w, "authorization header required")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims := &jwtClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				writeUnauthorized(w, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, &domain.AuthClaims{
				UserID:   claims.UserID,
				Nickname: claims.Nickname,
				IsAdmin:  claims.IsAdmin,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnly требует is_admin = true. Должен идти ПОСЛЕ Auth.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil || !claims.IsAdmin {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClaimsFromContext извлекает AuthClaims из контекста запроса.
func ClaimsFromContext(ctx context.Context) *domain.AuthClaims {
	v, _ := ctx.Value(ClaimsKey).(*domain.AuthClaims)
	return v
}

// jwtClaims — внутренняя структура для парсинга токена.
type jwtClaims struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
