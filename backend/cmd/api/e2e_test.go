//go:build e2e

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	"avito-internship-fs/internal/domain"
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

	server := newTestServer(ctx, t, nil)
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

	server := newTestServer(ctx, t, nil)
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

func TestE2E_CreatePendingRunHonorsConcurrentDeactivation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, nil)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	category := createCategory(t, server, adminToken, "Гонки", "")
	assistant := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId:   openapi_types.UUID(category.Id),
		Name:         "Спринтер",
		Description:  "проверка конкурентной деактивации",
		Model:        "gpt-4o-mini",
		SystemPrompt: "test",
	})

	tx, err := server.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin locking tx: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := tx.QueryRowContext(ctx, "SELECT id FROM assistants WHERE id = $1 FOR UPDATE", assistant.Id).Scan(new(uuid.UUID)); err != nil {
		t.Fatalf("lock assistant: %v", err)
	}

	runRepo := repository.NewRunRepository(server.db)
	resultCh := make(chan error, 1)
	go func() {
		runCtx, runCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer runCancel()
		_, _, err := runRepo.CreatePendingForActiveAssistant(runCtx, assistant.Id, auth.DummyUserID, "hello")
		resultCh <- err
	}()

	time.Sleep(150 * time.Millisecond)

	if _, err := tx.ExecContext(ctx, "UPDATE assistants SET is_active = FALSE WHERE id = $1", assistant.Id); err != nil {
		t.Fatalf("deactivate assistant in locking tx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit locking tx: %v", err)
	}

	err = <-resultCh
	if !errors.Is(err, domain.ErrAssistantInactive) {
		t.Fatalf("expected ErrAssistantInactive after concurrent deactivation, got %v", err)
	}

	var runsCount int
	if err := server.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assistant_runs WHERE assistant_id = $1", assistant.Id).Scan(&runsCount); err != nil {
		t.Fatalf("count assistant runs: %v", err)
	}
	if runsCount != 0 {
		t.Fatalf("expected no runs to be created, got %d", runsCount)
	}
}

// TestE2E_AssistantListFilteredByCategoryId verifies that GET /assistants?categoryId=...
// returns only assistants from the requested category. Hits the real SQL filter
// in the assistants repository via the full HTTP router.
func TestE2E_AssistantListFilteredByCategoryId(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, nil)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	userToken := dummyLogin(t, server, auth.RoleUser)

	foodCat := createCategory(t, server, adminToken, "Еда", "")
	sportCat := createCategory(t, server, adminToken, "Спорт", "")

	cook := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(foodCat.Id), Name: "Повар",
		Description: "Рецепты", Model: "gpt-4o-mini", SystemPrompt: "Ты повар",
	})
	baker := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(foodCat.Id), Name: "Пекарь",
		Description: "Выпечка", Model: "gpt-4o-mini", SystemPrompt: "Ты пекарь",
	})
	coach := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(sportCat.Id), Name: "Тренер",
		Description: "Тренировки", Model: "gpt-4o-mini", SystemPrompt: "Ты тренер",
	})

	resp := listAssistants(t, server, userToken, "?categoryId="+foodCat.Id.String())
	if resp.Pagination.Total != 2 {
		t.Fatalf("food filter: expected total=2, got %d", resp.Pagination.Total)
	}
	gotIDs := map[uuid.UUID]bool{}
	for _, a := range resp.Assistants {
		gotIDs[a.Id] = true
		if a.CategoryId != openapi_types.UUID(foodCat.Id) {
			t.Fatalf("food filter: assistant %s has categoryId=%s, expected %s", a.Id, a.CategoryId, foodCat.Id)
		}
	}
	if !gotIDs[cook.Id] || !gotIDs[baker.Id] {
		t.Fatalf("food filter: expected cook+baker, got %+v", resp.Assistants)
	}
	if gotIDs[coach.Id] {
		t.Fatalf("food filter: coach from sport category leaked in")
	}

	resp = listAssistants(t, server, userToken, "?categoryId="+sportCat.Id.String())
	if resp.Pagination.Total != 1 || len(resp.Assistants) != 1 || resp.Assistants[0].Id != coach.Id {
		t.Fatalf("sport filter: expected only coach, got %+v", resp.Assistants)
	}

	resp = listAssistants(t, server, userToken, "")
	if resp.Pagination.Total != 3 {
		t.Fatalf("no filter: expected total=3, got %d", resp.Pagination.Total)
	}
}

// TestE2E_UserFavoriteAssistants verifies the favorite_assistants relation through
// the public HTTP API: mark, read favorite flags, filter by favorites and unmark.
func TestE2E_UserFavoriteAssistants(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, nil)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	userToken := dummyLogin(t, server, auth.RoleUser)

	category := createCategory(t, server, adminToken, "Еда", "")
	cook := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(category.Id), Name: "Повар",
		Description: "Рецепты", Model: "gpt-4o-mini", SystemPrompt: "Ты повар",
	})
	baker := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(category.Id), Name: "Пекарь",
		Description: "Выпечка", Model: "gpt-4o-mini", SystemPrompt: "Ты пекарь",
	})

	resp := listAssistants(t, server, userToken, "")
	for _, a := range resp.Assistants {
		if a.IsFavorite {
			t.Fatalf("new assistants must not be favorite by default: %+v", a)
		}
	}

	setFavorite(t, server, userToken, cook.Id, true)

	status, raw := doJSON(t, server, http.MethodGet, "/assistants/"+cook.Id.String(), userToken, nil)
	if status != http.StatusOK {
		t.Fatalf("get favorite assistant: status=%d body=%s", status, raw)
	}
	var detail api.Assistant
	if err := json.Unmarshal(raw, &detail); err != nil {
		t.Fatalf("decode assistant: %v", err)
	}
	if !detail.IsFavorite {
		t.Fatalf("expected detail isFavorite=true")
	}
	if detail.SystemPrompt != nil {
		t.Fatalf("regular user must not see systemPrompt: %+v", detail.SystemPrompt)
	}

	resp = listAssistants(t, server, userToken, "?favoriteOnly=true")
	if resp.Pagination.Total != 1 || len(resp.Assistants) != 1 || resp.Assistants[0].Id != cook.Id {
		t.Fatalf("favorite filter: expected only cook, got %+v", resp)
	}
	if !resp.Assistants[0].IsFavorite {
		t.Fatalf("favorite filter item must have isFavorite=true")
	}

	resp = listAssistants(t, server, userToken, "")
	favoritesByID := map[uuid.UUID]bool{}
	for _, a := range resp.Assistants {
		favoritesByID[a.Id] = a.IsFavorite
	}
	if !favoritesByID[cook.Id] {
		t.Fatalf("cook should remain favorite in full list")
	}
	if favoritesByID[baker.Id] {
		t.Fatalf("baker should not be favorite")
	}

	setFavorite(t, server, userToken, cook.Id, false)
	resp = listAssistants(t, server, userToken, "?favoriteOnly=true")
	if resp.Pagination.Total != 0 || len(resp.Assistants) != 0 {
		t.Fatalf("favorite filter after removal: expected empty, got %+v", resp)
	}
}

// TestE2E_CategoriesListReturnsCreatedItems covers GET /categories after
// admin creates several categories — exercises the categories.List repository path.
func TestE2E_CategoriesListReturnsCreatedItems(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, nil)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	userToken := dummyLogin(t, server, auth.RoleUser)

	createCategory(t, server, adminToken, "Еда", "Рецепты")
	createCategory(t, server, adminToken, "Спорт", "Тренировки")

	status, raw := doJSON(t, server, http.MethodGet, "/categories", userToken, nil)
	if status != http.StatusOK {
		t.Fatalf("list categories: status=%d body=%s", status, raw)
	}
	var resp struct {
		Categories []api.Category `json:"categories"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(resp.Categories))
	}
}

// TestE2E_AssistantCreateRejectsNonexistentCategory verifies that creating an
// assistant with a categoryId that does not exist fails — exercises the
// categories.Exists check in the repository layer.
func TestE2E_AssistantCreateRejectsNonexistentCategory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, nil)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)

	body := api.AssistantCreateIn{
		CategoryId:   uuid.New(),
		Name:         "Призрак",
		Description:  "не должно создаться",
		Model:        "gpt-4o-mini",
		SystemPrompt: "test",
	}
	status, raw := doJSON(t, server, http.MethodPost, "/assistants", adminToken, body)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing category, got %d: %s", status, raw)
	}
	if !strings.Contains(string(raw), "CATEGORY_NOT_FOUND") {
		t.Fatalf("expected CATEGORY_NOT_FOUND code, got %s", raw)
	}
}

// TestE2E_AssistantUpdate verifies PUT /assistants/{id} writes new fields and
// can flip is_active — exercises assistants.Update in the repository.
func TestE2E_AssistantUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, nil)
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)

	category := createCategory(t, server, adminToken, "Еда", "")
	assistant := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(category.Id), Name: "Повар",
		Description: "первая версия", Model: "gpt-4o-mini", SystemPrompt: "v1",
	})

	updated := api.AssistantUpdateIn{
		CategoryId:   openapi_types.UUID(category.Id),
		Name:         "Повар pro",
		Description:  "обновлённая версия",
		Model:        "gpt-4o-mini",
		SystemPrompt: "v2",
		IsActive:     false,
	}
	status, raw := doJSON(t, server, http.MethodPut, "/assistants/"+assistant.Id.String(), adminToken, updated)
	if status != http.StatusOK {
		t.Fatalf("update: status=%d body=%s", status, raw)
	}
	var got api.Assistant
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Повар pro" || got.Description != "обновлённая версия" || got.IsActive {
		t.Fatalf("update did not persist: %+v", got)
	}
	if got.SystemPrompt == nil || *got.SystemPrompt != "v2" {
		t.Fatalf("systemPrompt not updated: %+v", got.SystemPrompt)
	}
}

// TestE2E_RunPersistsFailedOnProviderError verifies that when the LLM provider
// returns an error, the run is persisted with status=failed and shows up in /runs/my.
// Exercises runs.MarkFailed in the repository.
func TestE2E_RunPersistsFailedOnProviderError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	server := newTestServer(ctx, t, &failingProvider{msg: "synthetic provider failure"})
	t.Cleanup(server.Close)

	adminToken := dummyLogin(t, server, auth.RoleAdmin)
	userToken := dummyLogin(t, server, auth.RoleUser)

	category := createCategory(t, server, adminToken, "Еда", "")
	assistant := createAssistant(t, server, adminToken, api.AssistantCreateIn{
		CategoryId: openapi_types.UUID(category.Id), Name: "Повар",
		Description: "падающий", Model: "gpt-4o-mini", SystemPrompt: "fail me",
	})

	status, raw := doJSON(t, server, http.MethodPost,
		"/assistants/"+assistant.Id.String()+"/run", userToken,
		api.AssistantRunCreateIn{UserPrompt: "hi"})
	if status != http.StatusBadGateway {
		t.Fatalf("expected 502 for provider error, got %d: %s", status, raw)
	}

	myRuns := listMyRuns(t, server, userToken)
	if myRuns.Pagination.Total != 1 || len(myRuns.Runs) != 1 {
		t.Fatalf("expected 1 persisted failed run, got %+v", myRuns)
	}
	if myRuns.Runs[0].Status != api.Failed {
		t.Fatalf("expected status=failed, got %s", myRuns.Runs[0].Status)
	}
	if myRuns.Runs[0].Error == nil || !strings.Contains(*myRuns.Runs[0].Error, "synthetic provider failure") {
		t.Fatalf("expected error to mention provider failure, got %+v", myRuns.Runs[0].Error)
	}
}

type failingProvider struct {
	msg string
}

func (f *failingProvider) Generate(_ context.Context, _ llm.Request) (llm.Response, error) {
	return llm.Response{}, errors.New(f.msg)
}

func (f *failingProvider) GenerateStream(_ context.Context, _ llm.Request, _ func(llm.StreamChunk)) (llm.Response, error) {
	return llm.Response{}, errors.New(f.msg)
}

type testServer struct {
	httpSrv   *httptest.Server
	db        *sql.DB
	pgCleanup func()
}

func (s *testServer) Close() {
	s.httpSrv.Close()
	if s.pgCleanup != nil {
		s.pgCleanup()
	}
}

func (s *testServer) URL() string { return s.httpSrv.URL }

func newTestServer(ctx context.Context, t *testing.T, provider llm.Provider) *testServer {
	t.Helper()

	if provider == nil {
		provider = llm.NewMockProvider()
	}

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
	userRepo := repository.NewUserRepository(db)

	handler := newRouter(routerDeps{
		Issuer:            issuer,
		AuthHandler:       auth.NewHandler(issuer, userRepo),
		CategoriesHandler: categories.NewHandler(service.NewCategoryService(categoryRepo)),
		AssistantsHandler: assistants.NewHandler(service.NewAssistantService(assistantRepo)),
		RunsHandler:       runs.NewHandler(service.NewRunService(runRepo, provider, 5*time.Second)),
	})

	httpSrv := httptest.NewServer(handler)

	return &testServer{
		httpSrv: httpSrv,
		db:      db,
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

func setFavorite(t *testing.T, s *testServer, token string, assistantID uuid.UUID, favorite bool) {
	t.Helper()
	method := http.MethodPut
	if !favorite {
		method = http.MethodDelete
	}
	status, raw := doJSON(t, s, method, "/assistants/"+assistantID.String()+"/favorite", token, nil)
	if status != http.StatusNoContent {
		t.Fatalf("set favorite=%v: status=%d body=%s", favorite, status, raw)
	}
}

type assistantListResp struct {
	Assistants []api.Assistant `json:"assistants"`
	Pagination api.Pagination  `json:"pagination"`
}

func listAssistants(t *testing.T, s *testServer, token, query string) assistantListResp {
	t.Helper()
	status, raw := doJSON(t, s, http.MethodGet, "/assistants"+query, token, nil)
	if status != http.StatusOK {
		t.Fatalf("list assistants %q: status=%d body=%s", query, status, raw)
	}
	var resp assistantListResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode assistants: %v", err)
	}
	return resp
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
