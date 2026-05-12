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

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			reqHeaders := r.Header.Get("Access-Control-Request-Headers")
			if reqHeaders == "" {
				reqHeaders = "Authorization, Content-Type"
			}
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
