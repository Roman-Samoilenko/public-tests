package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"quiz-platform/internal/config"
	"quiz-platform/internal/handler"
	pg "quiz-platform/internal/repository/postgres"

	"golang.org/x/net/publicsuffix"
)

// newIPv4OnlyHTTPClient создаёт HTTP-клиент, который работает только через IPv4.
func newIPv4OnlyHTTPClient() *http.Client {
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: false,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		IdleConnTimeout:     30 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
		Jar:       jar,
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	cfg := config.Load()

	ctx := context.Background()
	db, err := pg.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	profileRepo := pg.NewProfileRepository(db)
	answerRepo := pg.NewAnswerRepository(db)
	testRepo := pg.NewTestRepository(db)

	profileHandler := handler.NewProfileHandler(profileRepo, answerRepo)
	testHandler := handler.NewTestHandler(testRepo, answerRepo)
	adminHandler := handler.NewAdminHandler(testRepo)

	httpClient := newIPv4OnlyHTTPClient()
	importHandler := handler.NewImportHandler(httpClient)

	router := handler.NewRouter(profileHandler, importHandler, testHandler, adminHandler, cfg.JWTSecret)

	addr := ":" + cfg.HTTPPort
	slog.Info("main service starting", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
