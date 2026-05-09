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
	"avito-internship-fs/internal/database"
	"avito-internship-fs/internal/httpx"
)

func main() {
	addr := env("HTTP_ADDR", ":8080")
	dbURL := env("DATABASE_URL", "postgres://assistants:assistants@localhost:5432/assistants?sslmode=disable")
	jwtSecret := env("JWT_SECRET", "dev-secret-change-me")

	db, err := database.Connect(dbURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	_ = db

	issuer, err := auth.NewIssuer(jwtSecret, 24*time.Hour)
	if err != nil {
		slog.Error("auth issuer init failed", "error", err)
		os.Exit(1)
	}
	authHandler := auth.NewHandler(issuer)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("starting backend", "addr", addr)
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

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
