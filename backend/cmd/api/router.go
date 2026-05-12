package main

import (
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"os"
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
	mux.HandleFunc("GET /docs", swaggerUI)
	mux.HandleFunc("GET /docs/", swaggerUI)
	mux.HandleFunc("GET /docs/openapi.yaml", openAPISpec)
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

var openAPISpecPaths = []string{
	"api.yaml",
	"../api.yaml",
	"../../api.yaml",
	"../../../api.yaml",
	"/api.yaml",
}

func openAPISpec(w http.ResponseWriter, _ *http.Request) {
	data, err := readOpenAPISpec()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "openapi spec is unavailable")
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func readOpenAPISpec() ([]byte, error) {
	var lastErr error
	for _, path := range openAPISpecPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			lastErr = err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fs.ErrNotExist
}

func swaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = swaggerUITemplate.Execute(w, nil)
}

var swaggerUITemplate = template.Must(template.New("swagger").Parse(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>AI Assistants Catalog API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; background: #fff; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/docs/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        persistAuthorization: true,
        displayRequestDuration: true
      });
    };
  </script>
</body>
</html>`))
