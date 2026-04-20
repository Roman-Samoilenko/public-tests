package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	authmw "quiz-platform/internal/middleware"
)

func NewRouter(
	profile *ProfileHandler,
	imports *ImportHandler,
	tests *TestHandler,
	admin *AdminHandler,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(authmw.Auth(jwtSecret))

		// Профиль текущего пользователя
		r.Get("/profile", profile.GetProfile)
		r.Put("/profile", profile.UpdateProfile)
		r.Get("/profile/answers", profile.GetAnswerHistory)

		// Импорт тестов из внешних источников
		r.Post("/import/google-forms", imports.ImportGoogleForm)

		// Тесты
		r.Route("/tests", func(r chi.Router) {
			r.Get("/", tests.ListTests)
			r.Post("/", tests.CreateTest)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/my-answer", tests.GetMyAnswer)
				r.Get("/", tests.GetTest)
				r.Post("/answers", tests.SubmitAnswer)
				r.Post("/vote", tests.VoteTest)
				r.Get("/comments", tests.GetComments)
				r.Post("/comments", tests.AddComment)
				r.Delete("/comments/{commentId}", tests.DeleteComment)
			})
		})

		// Админ
		r.Route("/admin", func(r chi.Router) {
			r.Use(authmw.AdminOnly)
			r.Post("/tests/{id}/official", admin.MakeOfficial)
		})
	})

	return r
}
