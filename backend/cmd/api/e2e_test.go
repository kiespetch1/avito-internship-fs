//go:build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/assistants"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/categories"
	"avito-internship-fs/internal/database"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/repository"
	"avito-internship-fs/internal/runs"
	"avito-internship-fs/internal/service"
)

// TestE2E_AdminCreatesCategoryAndAssistant_UserRunsIt covers the assignment-required
// end-to-end scenario:
//  1. admin creates a category
//  2. admin creates an assistant in that category
//  3. user runs the assistant
//  4. the run shows up in /runs/my
//
// Uses a real Postgres in a testcontainer with all migrations applied,
// the real HTTP router and the mock LLM provider.
func TestE2E_AdminCreatesCategoryAndAssistant_UserRunsIt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	userToken := dummyLogin(t, server, auth.RoleUser)

	category := createCategory(t, server, adminToken, "Еда", "Рецепты и кулинария")
	assistant := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId:   openapi_types.UUID(category.Id),
		Name:         "Повар",
		Description:  "Составляет рецепты из ингредиентов",
		Model:        "gpt-4o-mini",
		SystemPrompt: "Ты опытный повар",
	})
	if !assistant.IsActive {
		t.Fatalf("expected new assistant to be active by default")
	}

	run := runAssistant(t, server, userToken, assistant.Id, "курица, рис, томаты")
	if run.Status != api.Success {
		t.Fatalf("expected run status=success, got %s (error=%v)", run.Status, derefStr(run.Error))
	}
	if run.UserId != auth.DummyUserID {
		t.Fatalf("expected run userId=%s, got %s", auth.DummyUserID, run.UserId)
	}
	if got := derefStr(run.Output); !strings.Contains(got, "курица") {
		t.Fatalf("expected mock provider output to echo user prompt, got %q", got)
	}

	myRuns := listMyRuns(t, server, userToken)
	if myRuns.Pagination.Total != 1 {
		t.Fatalf("expected /runs/my total=1, got %d", myRuns.Pagination.Total)
	}
	if len(myRuns.Runs) != 1 || myRuns.Runs[0].Id != run.Id {
		t.Fatalf("expected /runs/my to contain the new run %s, got %+v", run.Id, myRuns.Runs)
	}
}

// TestE2E_UserCannotRunInactiveAssistant verifies the business rule that
// an inactive assistant cannot be executed.
func TestE2E_UserCannotRunInactiveAssistant(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	userToken := dummyLogin(t, server, auth.RoleUser)

	category := createCategory(t, server, adminToken, "Спорт", "")
	isActive := false
	assistant := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId:   openapi_types.UUID(category.Id),
		Name:         "Тренер",
		Description:  "План тренировок",
		Model:        "gpt-4o-mini",
		SystemPrompt: "Ты тренер",
		IsActive:     &isActive,
	})

	status, body := doJSON(t, server, http.MethodPost,
		"/assistants/"+assistant.Id.String()+"/run",
		userToken, api.AssistantRunCreateIn{UserPrompt: "5 км бег"},
	)
	if status != http.StatusConflict {
		t.Fatalf("expected 409 for inactive assistant run, got %d: %s", status, body)
	}
}

// ----- helpers -----

type testServer struct {
	httpSrv   *httptest.Server
	pgCleanup func()
}

func (s *testServer) Close() {
	s.httpSrv.Close()
	if s.pgCleanup != nil {
		s.pgCleanup()
	}
}

func (s *testServer) URL() string { return s.httpSrv.URL }

func newTestServer(ctx context.Context, t *testing.T) *testServer {
	t.Helper()

	pgC, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ai_assistants_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	connStr, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(pgC)
		t.Fatalf("get postgres connection string: %v", err)
	}

	db, err := database.Connect(connStr)
	if err != nil {
		_ = testcontainers.TerminateContainer(pgC)
		t.Fatalf("connect+migrate db: %v", err)
	}

	issuer, err := auth.NewIssuer("e2e-test-secret-please-change", time.Hour)
	if err != nil {
		_ = db.Close()
		_ = testcontainers.TerminateContainer(pgC)
		t.Fatalf("issuer init: %v", err)
	}

	categoryRepo := repository.NewCategoryRepository(db)
	assistantRepo := repository.NewAssistantRepository(db)
	runRepo := repository.NewRunRepository(db)

	handler := newRouter(routerDeps{
		Issuer:            issuer,
		AuthHandler:       auth.NewHandler(issuer),
		CategoriesHandler: categories.NewHandler(service.NewCategoryService(categoryRepo)),
		AssistantsHandler: assistants.NewHandler(service.NewAssistantService(assistantRepo)),
		RunsHandler: runs.NewHandler(service.NewRunService(
			assistantRepo, runRepo, llm.NewMockProvider(), 5*time.Second,
		)),
	})

	httpSrv := httptest.NewServer(handler)

	return &testServer{
		httpSrv: httpSrv,
		pgCleanup: func() {
			_ = db.Close()
			_ = testcontainers.TerminateContainer(pgC)
		},
	}
}

func dummyLogin(t *testing.T, s *testServer, role auth.Role) string {
	t.Helper()
	var apiRole api.Role
	switch role {
	case auth.RoleAdmin:
		apiRole = api.RoleAdmin
	case auth.RoleUser:
		apiRole = api.RoleUser
	default:
		t.Fatalf("unknown role %v", role)
	}
	var resp api.Token
	status, body := doJSON(t, s, http.MethodPost, "/dummyLogin", "",
		map[string]string{"role": string(apiRole)})
	if status != http.StatusOK {
		t.Fatalf("dummyLogin %s: status=%d body=%s", apiRole, status, body)
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("dummyLogin returned empty token")
	}
	return resp.Token
}

func createCategory(t *testing.T, s *testServer, token, name, description string) api.Category {
	t.Helper()
	desc := description
	body := api.CategoryCreateIn{Name: name, Description: &desc}
	status, raw := doJSON(t, s, http.MethodPost, "/categories", token, body)
	if status != http.StatusCreated {
		t.Fatalf("create category: status=%d body=%s", status, raw)
	}
	var c api.Category
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("decode category: %v", err)
	}
	if c.Id == (uuid.UUID{}) {
		t.Fatalf("create category returned empty id")
	}
	return c
}

func createAssistant(t *testing.T, s *testServer, token string, in api.AssistantCreateIn) api.Assistant {
	t.Helper()
	status, raw := doJSON(t, s, http.MethodPost, "/assistants", token, in)
	if status != http.StatusCreated {
		t.Fatalf("create assistant: status=%d body=%s", status, raw)
	}
	var a api.Assistant
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatalf("decode assistant: %v", err)
	}
	return a
}

func runAssistant(t *testing.T, s *testServer, token string, assistantID uuid.UUID, prompt string) api.AssistantRun {
	t.Helper()
	status, raw := doJSON(t, s, http.MethodPost,
		"/assistants/"+assistantID.String()+"/run", token,
		api.AssistantRunCreateIn{UserPrompt: prompt},
	)
	if status != http.StatusCreated {
		t.Fatalf("run assistant: status=%d body=%s", status, raw)
	}
	var r api.AssistantRun
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("decode run: %v", err)
	}
	return r
}

type runListResp struct {
	Runs       []api.AssistantRun `json:"runs"`
	Pagination api.Pagination     `json:"pagination"`
}

func listMyRuns(t *testing.T, s *testServer, token string) runListResp {
	t.Helper()
	status, raw := doJSON(t, s, http.MethodGet, "/runs/my", token, nil)
	if status != http.StatusOK {
		t.Fatalf("list my runs: status=%d body=%s", status, raw)
	}
	var resp runListResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode my runs: %v", err)
	}
	return resp
}

func doJSON(t *testing.T, s *testServer, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, s.URL()+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, raw
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
