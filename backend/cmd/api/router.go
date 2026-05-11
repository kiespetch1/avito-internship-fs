package main

import (
	"net/http"
	"time"

	"avito-internship-fs/internal/assistants"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/categories"
	"avito-internship-fs/internal/httpx"
	"avito-internship-fs/internal/runs"
)

type routerDeps struct {
	Issuer            *auth.Issuer
	AuthHandler       *auth.Handler
	CategoriesHandler *categories.Handler
	AssistantsHandler *assistants.Handler
	RunsHandler       *runs.Handler
}

func newRouter(d routerDeps) http.Handler {
	authed := auth.RequireAuth(d.Issuer)
	adminOnly := func(h http.Handler) http.Handler {
		return authed(auth.RequireRole(auth.RoleAdmin)(h))
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /_info", healthcheck)
	mux.HandleFunc("POST /dummyLogin", d.AuthHandler.DummyLogin)

	mux.Handle("GET /categories", authed(http.HandlerFunc(d.CategoriesHandler.List)))
	mux.Handle("POST /categories", adminOnly(http.HandlerFunc(d.CategoriesHandler.Create)))

	mux.Handle("GET /assistants", authed(http.HandlerFunc(d.AssistantsHandler.List)))
	mux.Handle("POST /assistants", adminOnly(http.HandlerFunc(d.AssistantsHandler.Create)))
	mux.Handle("GET /assistants/{assistantId}", authed(http.HandlerFunc(d.AssistantsHandler.Get)))
	mux.Handle("PUT /assistants/{assistantId}", adminOnly(http.HandlerFunc(d.AssistantsHandler.Update)))
	mux.Handle("POST /assistants/{assistantId}/run", authed(http.HandlerFunc(d.RunsHandler.Run)))

	mux.Handle("GET /runs/my", authed(http.HandlerFunc(d.RunsHandler.MyRuns)))
	mux.Handle("GET /admin/runs", adminOnly(http.HandlerFunc(d.RunsHandler.AdminRuns)))

	return mux
}

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
