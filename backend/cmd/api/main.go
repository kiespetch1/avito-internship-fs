package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"avito-internship-fs/internal/assistants"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/categories"
	"avito-internship-fs/internal/config"
	"avito-internship-fs/internal/database"
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

	provider, err := resolveLLMProvider(cfg.LLM)
	if err != nil {
		slog.Error("llm provider init failed", "error", err)
		os.Exit(1)
	}

	categoryRepo := repository.NewCategoryRepository(db)
	assistantRepo := repository.NewAssistantRepository(db)
	runRepo := repository.NewRunRepository(db)
	userRepo := repository.NewUserRepository(db)

	handler := newRouter(routerDeps{
		Issuer:            issuer,
		AuthHandler:       auth.NewHandler(issuer, userRepo),
		CategoriesHandler: categories.NewHandler(service.NewCategoryService(categoryRepo)),
		AssistantsHandler: assistants.NewHandler(service.NewAssistantService(assistantRepo)),
		RunsHandler:       runs.NewHandler(service.NewRunService(runRepo, provider, cfg.LLM.Timeout)),
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("starting backend",
			"addr", cfg.HTTPAddr,
			"llm.provider", cfg.LLM.Provider,
			"llm.timeout", cfg.LLM.Timeout,
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
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "mock", "":
		return llm.NewMockProvider(), nil
	case "openai", "openai-compatible", "openai_compatible":
		baseURL := strings.TrimSpace(cfg.BaseURL)
		if baseURL == "" {
			baseURL = llm.DefaultOpenAIBaseURL
		}

		return llm.NewOpenAICompatibleProvider(llm.OpenAICompatibleConfig{
			BaseURL: baseURL,
			APIKey:  cfg.APIKey,
		})
	default:
		return nil, errors.New("unsupported llm provider: " + cfg.Provider)
	}
}
