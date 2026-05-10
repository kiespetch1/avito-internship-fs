package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/categories"
	"avito-internship-fs/internal/config"
	"avito-internship-fs/internal/database"
	"avito-internship-fs/internal/httpx"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/repository"
	"avito-internship-fs/internal/runs"
	"avito-internship-fs/internal/service"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	issuer, err := auth.NewIssuer(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		slog.Error("auth issuer init failed", "error", err)
		os.Exit(1)
	}
	authHandler := auth.NewHandler(issuer)

	provider, err := resolveLLMProvider(cfg.LLM)
	if err != nil {
		slog.Error("llm provider init failed", "error", err)
		os.Exit(1)
	}

	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoriesHandler := categories.NewHandler(categoryService)

	assistantRepo := repository.NewAssistantRepository(db)
	runRepo := repository.NewRunRepository(db)
	runService := service.NewRunService(assistantRepo, runRepo, provider, cfg.LLM.Timeout)
	runsHandler := runs.NewHandler(runService)

	authed := auth.RequireAuth(issuer)
	adminOnly := func(h http.Handler) http.Handler {
		return authed(auth.RequireRole(auth.RoleAdmin)(h))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)
	mux.Handle("GET /categories", authed(http.HandlerFunc(categoriesHandler.List)))
	mux.Handle("POST /categories", adminOnly(http.HandlerFunc(categoriesHandler.Create)))
	mux.Handle("POST /assistants/{assistantId}/run", authed(http.HandlerFunc(runsHandler.Run)))

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("starting backend",
			"addr", cfg.HTTPAddr,
			"llm.provider", cfg.LLM.Provider,
			"llm.timeout", cfg.LLM.Timeout,
			"llm.model", cfg.LLM.DefaultModel,
			"jwt.ttl", cfg.JWTTTL,
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}

	_ = db.Close()
	slog.Info("backend stopped")
}

func resolveLLMProvider(cfg config.LLMConfig) (llm.Provider, error) {
	switch cfg.Provider {
	case "mock", "":
		return llm.NewMockProvider(), nil
	default:
		return nil, errors.New("unsupported llm provider: " + cfg.Provider)
	}
}
