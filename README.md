# Каталог AI-ассистентов

Fullstack-приложение для управления, просмотра и запуска AI-ассистентов.

## Запуск

```bash
docker compose up --build
```

- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- Healthcheck: `GET http://localhost:8080/_info`
- Swagger UI: http://localhost:8080/docs
- OpenAPI YAML: http://localhost:8080/docs/openapi.yaml

## Конфигурация

Backend читает настройки в порядке (от низшего приоритета к высшему):

1. `backend/config.yaml` (опционально) — рядом с исполняемым файлом. Путь переопределяется через `CONFIG_PATH`.
2. Переменные окружения — имеют приоритет над значениями из yaml-конфига.

Пример `config.yaml`:

```yaml
httpAddr: ":8080"
databaseUrl: "postgres://assistants:assistants@localhost:5432/assistants?sslmode=disable"
jwtSecret: "dev-secret-change-me"
jwtTtl: 24h

llm:
  provider: mock
  timeout: 120s
  defaultModel: gpt-mock
  baseUrl: "" # для openai-compatible по умолчанию будет https://api.openai.com/v1
  apiKey: ""  # только через env для реального провайдера
```

Значения по умолчанию задаются только в одном месте — в `docker-compose.yaml`.

### Почему env-переменные приоритетнее yaml-конфига
1. Можно быстро поменять параметр без правки файлов и пересборки образа.
2. Задание прямо требует, чтобы `LLM_API_KEY` для реального провайдера читался только из env. Если бы yaml перебивал env, секрет в файле мог бы перетереть значение из secret manager.


### Переменные

| Переменная           | Дефолт в compose                                                              | Назначение                                       |
| -------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------ |
| `HTTP_ADDR`          | `:8080`                                                                       | Адрес HTTP-сервера                               |
| `DATABASE_URL`       | `postgres://assistants:assistants@postgres:5432/assistants?sslmode=disable`   | Подключение к PostgreSQL                         |
| `JWT_SECRET`         | `dev-secret`                                                                  | Секрет для подписи JWT                           |
| `JWT_TTL`            | `24h`                                                                         | Время жизни JWT (`time.Duration`)                |
| `LLM_PROVIDER`       | `mock`                                                                        | Идентификатор LLM-провайдера: `mock`, `openai-compatible`, `openai` |
| `LLM_TIMEOUT`        | `120s`                                                                        | Таймаут запроса к LLM                            |
| `LLM_DEFAULT_MODEL`  | `gpt-mock`                                                                    | Модель по умолчанию, если ассистент не задал свою |
| `LLM_BASE_URL`       | —                                                                             | URL OpenAI-совместимого провайдера; если не задан, используется `https://api.openai.com/v1` |
| `LLM_API_KEY`        | —                                                                             | API-ключ внешнего провайдера (только через env)  |
| `CONFIG_PATH`        | —                                                                             | Путь к yaml-конфигу (по умолчанию `config.yaml`) |

## Архитектурные решения

### Схема БД и индексы

Схема состоит из семи таблиц (миграции `0001`–`0007` в `backend/internal/database/migrations/`):

| Таблица              | Назначение                                             |
| -------------------- |--------------------------------------------------------|
| `users`              | Пользователи, роль `admin` / `user`, bcrypt-хэш пароля |
| `categories`         | Категории ассистентов                                  |
| `assistants`         | Карточка ассистента: модель, промпты, флаг `is_active` |
| `assistant_runs`     | История запусков: статус, токены, latency              |
| `tags`               | Словарь тегов                                          |
| `assistant_tags`     | M2M: отношение ассистента к тегу                       |
| `favorite_assistants`| M2M: отношение пользователя к ассистенту (избранному)  |
| `run_feedback`       | Лайк/дизлайк (`rating IN (-1, 1)`) на запуск           |

**Индексы:**

```sql
-- Каталог: фильтрация по категории и статусу активности
CREATE INDEX idx_assistants_category_id ON assistants (category_id);
CREATE INDEX idx_assistants_active      ON assistants (is_active) WHERE is_active = TRUE;

-- Полнотекстовый поиск (ILIKE '%q%') через pg_trgm
CREATE INDEX idx_assistants_name_trgm ON assistants USING GIN (name gin_trgm_ops);
CREATE INDEX idx_assistants_desc_trgm ON assistants USING GIN (description gin_trgm_ops);

-- История запусков: сортировка по дате, фильтрация по статусу и ассистенту
CREATE INDEX idx_runs_user_created ON assistant_runs (user_id, created_at DESC);
CREATE INDEX idx_runs_status       ON assistant_runs (status);
CREATE INDEX idx_runs_assistant    ON assistant_runs (assistant_id);

-- Поиск ассистентов по тегу
CREATE INDEX idx_assistant_tags_tag ON assistant_tags (tag_id);
```

### Регистрация и авторизация

Есть два независимых способа получить JWT:

| Endpoint | Назначение |
| --- | --- |
| `POST /auth/register` | Создаёт нового пользователя с ролью `user` и сразу возвращает JWT |
| `POST /auth/login` | Авторизует существующего пользователя по email и паролю |
| `POST /dummyLogin` | Тестовый вход без пароля от имени `admin` или `user` |

Регистрация администратора через публичный API запрещена: endpoint регистрации не принимает поле `role`, JSON декодируется с `DisallowUnknownFields`, а repository всегда вставляет роль `'user'`. Тестовые пользователи для `/dummyLogin` остаются сидом миграции `0001` и нужны только для проверки ролей.

Пароли не хранятся в открытом виде. Backend нормализует email (`trim` + lowercase), валидирует пароль на сервере и сохраняет только bcrypt-хэш с серверным cost. Клиент не может выбрать алгоритм хэширования или передать соль: эти параметры задаются только backend-кодом. Для bcrypt дополнительно ограничена длина пароля до 72 байт, чтобы не попасть в неочевидное усечение.

JWT подписывается только `HS256`; при разборе токена backend отклоняет любой другой `alg`, проверяет срок действия, UUID пользователя и роль из enum `admin` / `user`. Для неверного email и неверного пароля `POST /auth/login` возвращает одинаковый `401 UNAUTHORIZED`, чтобы не раскрывать существование аккаунта через сообщение об ошибке.

### Обработка запуска ассистента и сохранение истории

Реализовано в `backend/internal/service/runs.go`. Запуск проходит через несколько контекстов с разными гарантиями:

1. **Проверка ассистента** — в рамках HTTP-запроса.
2. **Сохранение `pending`** — `context.Background()` + 5 с. Запись происходит до вызова LLM; если БД недоступна — HTTP 500 без утечки.
3. **Вызов LLM для обычного `POST /assistants/{assistantId}/run`** — `context.Background()` + `LLM_TIMEOUT` (120 с по умолчанию). Контекст не привязан к HTTP-запросу: если клиент отключится, backend всё равно доведёт уже созданный запуск до финального состояния.
4. **Вызов LLM для streaming `POST /assistants/{assistantId}/run/stream`** — контекст HTTP-запроса + `LLM_TIMEOUT`. Если клиент закрывает вкладку, отменяет `fetch` или запись SSE ломается, backend отменяет upstream LLM-вызов.
5. **Финализация** — ещё один `context.Background()` + 5 с. Запуск переводится в `success` или `failed` независимо от жизни соединения.

Финализация всегда использует отдельный короткий DB-контекст, поэтому запись в `assistant_runs` не должна зависать в `pending` из-за разрыва клиентского соединения.

### Контракт мокируемого LLM-провайдера

Интерфейс в `backend/internal/llm/provider.go`:

```go
type Request struct {
    Model        string
    SystemPrompt string
    UserPrompt   string
}

type Response struct {
    Output       string
    TokensIn     int
    TokensOut    int
    LatencyMs    int
    FinishReason string
}

type StreamChunk struct {
    Delta string
}

type Provider interface {
    Generate(ctx context.Context, req Request) (Response, error)
    GenerateStream(ctx context.Context, req Request, onChunk func(StreamChunk)) (Response, error)
}
```

`MockProvider` (`llm/mock.go`) реализует интерфейс без сетевых вызовов:
- возвращает `[mock:<model>] <userPrompt>` как `Output`;
- в streaming-режиме отдаёт тот же ответ частями через `onChunk`;
- оценивает токены как `len(text)/4`;

### Пример запроса к LLM

При вызове ассистента сервис формирует структуру `llm.Request` со значениями из БД:

```json
{
  "model": "gpt-4o",
  "systemPrompt": "Ты — ассистент по подбору жилья. Отвечай кратко и по делу.",
  "userPrompt": "Какие районы Москвы подходят для семьи с детьми?"
}
```

`systemPrompt` берётся из поля `assistants.system_prompt` (всегда `NOT NULL` в БД). `model` — из `assistants.model`, задаётся при создании ассистента администратором.

### OpenAI-compatible LLM-провайдер

Помимо `mock`, backend поддерживает внешний OpenAI-compatible Chat Completions API через `llm.OpenAICompatibleProvider`. Провайдер отправляет `systemPrompt` и `userPrompt` как `messages` в `POST /chat/completions`, читает `choices[0].message.content`, `usage.prompt_tokens`, `usage.completion_tokens` и `finish_reason`.

Для отображения прогресса генерации есть отдельный endpoint `POST /assistants/{assistantId}/run/stream`. Он создаёт запуск в статусе `pending`, отдаёт Server-Sent Events (`run`, `delta`, `done`, `failed`) и после завершения сохраняет финальный output в `assistant_runs`. Обычный `POST /assistants/{assistantId}/run` остаётся синхронным JSON-endpoint.

Streaming-режим OpenAI-compatible provider считает ответ успешным только после upstream-события `data: [DONE]`. Если соединение с провайдером закрывается после частичных чанков без `[DONE]`, запуск помечается как `failed`, чтобы не сохранять обрезанный ответ как `success`.

Интеграция реализована как OpenAI-compatible слой, а не как привязка к одному вендору. OpenAI API работает без `LLM_BASE_URL`, а, например, OpenRouter, подключается той же веткой через переопределённый base URL. Mock-провайдер остается дефолтом для проверки задания без ключей и позволяет включить реальный провайдер только env-переменными.

Если `LLM_BASE_URL` не задан, используется базовый OpenAI endpoint `https://api.openai.com/v1`:

```bash
LLM_PROVIDER=openai-compatible
LLM_API_KEY=sk-...
LLM_DEFAULT_MODEL=gpt-4o-mini
```

Для OpenRouter:

```bash
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_API_KEY=sk-or-...
LLM_DEFAULT_MODEL=openai/gpt-4o-mini
```

Реальный ключ не нужен для запуска проекта: дефолтный `LLM_PROVIDER=mock` остаётся полностью автономным.

### Стейт-менеджмент на frontend

Вся серверная асинхронная логика управляется **TanStack Query**:
- данные кешируются на 30 с (`staleTime: 30_000`), `refetchOnWindowFocus` отключён;
- повторные попытки не выполняются для 401 / 403 / 404, для остальных ошибок — до 2 повторов;
- глобальный обработчик в `QueryCache` / `MutationCache` разлогинивает пользователя при 401.

Стейт аутентификации (JWT-токен + профиль пользователя) живёт в `localStorage` под ключом `avito.auth.v1` и читается через **`useSyncExternalStore`** (без дополнительных библиотек, поскольку приложение не предполагает сложного стейта). Изменения распространяются на все открытые вкладки через `window.addEventListener('storage', ...)`. Хранилище — `frontend/src/lib/auth/storage.ts`. Вход, регистрация и `dummyLogin` выполняются через TanStack Query mutations (`frontend/src/lib/api/queries/useAuthMutations.ts`), после успешного ответа сохраняется единый session snapshot.

Формы управляются через **TanStack Form** + **Zod** (валидация схем в `frontend/src/lib/forms/schemas/`). Форма входа по умолчанию использует email/пароль; ниже доступно переключение на `/dummyLogin`, которое показывает прежний выбор роли `admin` / `user`. Кнопка регистрации находится на экране входа и переключает форму на создание пользовательского аккаунта.

### `systemPrompt` скрыт от обычного пользователя

В контракте `Assistant.systemPrompt` помечен `nullable: true`. Backend возвращает реальное значение только для роли `admin`, для роли `user` поле сериализуется как `null`.

Системный промпт — внутренняя настройка ассистента, в большинстве случаев пользователю не стоит ее знать, как минимум из-за находящихся там формулировок, ограничений, правил, в том числе ограничивающих обходы и злоупотребление. В БД поле хранится как `NOT NULL` — на стороне сервиса оно всегда есть.

## База данных и миграции

Миграции лежат в `backend/internal/database/migrations/`. Применяются автоматически через `golang-migrate` при старте backend, до открытия HTTP-порта. Если миграция падает, процесс завершается с ошибкой, сервер не стартует.

Применённая версия фиксируется в служебной таблице `schema_migrations`. Повторный запуск идемпотентен: уже применённые миграции пропускаются.

### Добавление миграции

Для добавения миграции необходимо положить два файла в `backend/internal/database/migrations/`:

```
NNNN_short_description.up.sql
NNNN_short_description.down.sql
```

Номер `NNNN` — следующий по порядку от последнего. Пересборка backend подхватит новые миграции автоматически.

### Healthcheck

- **Postgres**: `pg_isready` через docker healthcheck.
- **Backend**: `GET /_info` через docker healthcheck.
- `frontend` стартует только после готовности backend и БД.

Статус всех сервисов локально — `docker compose ps`.

### Swagger-документация

Backend отдаёт Swagger UI на `GET /docs`. Интерфейс использует корневой `api.yaml`, доступный как `GET /docs/openapi.yaml`, поэтому документация и контракт API остаются синхронизированы с OpenAPI-спецификацией.

Для `POST /assistants/{assistantId}/run/stream` в OpenAPI описаны как сам `text/event-stream` ответ (`200`), так и JSON-ошибки до старта SSE-потока: `400`, `401`, `404`, `409`, `500`, `502`. После отправки первого SSE-события HTTP-статус уже зафиксирован как `200`, поэтому ошибки генерации передаются событием `failed`.

## Тесты и покрытие

### Backend

Все команды запускаются из корня репозитория.

| Команда | Что делает |
| --- | --- |
| `make -C backend build` | Сборка бинарника `backend/bin/api` |
| `make -C backend lint` | `go vet` + `golangci-lint run ./...` (если установлен) |
| `make -C backend test` | Unit-тесты `go test ./...` |
| `make -C backend test-e2e` | E2E-тесты за тегом `e2e` через testcontainers (поднимает реальный Postgres в Docker) |

Подробный coverage:

```bash
cd backend
go test -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out      # текстовая сводка по функциям
go tool cover -html=coverage.out      # HTML-отчёт в браузере
```

Покрытие unit-тестами на момент написания (`go test -covermode=atomic ./...`):

| Пакет | Покрытие |
| --- | --- |
| `internal/httpx` | 94.3% |
| `internal/service` | 92.0% |
| `internal/llm` | 91.7% |
| `internal/runs` | 88.2% |
| `internal/assistants` | 86.7% |
| `internal/auth` | 85.2% |
| `internal/config` | 84.4% |
| `internal/categories` | 84.0% |
| `cmd/api` | 35.8% (бóльшая часть покрыта e2e за тегом) |
| `internal/database`, `internal/repository` | 0% (тонкие обёртки над `pgx`, покрываются e2e) |
| **Total** | **58.4%** |

Доменная логика (сервисы, валидация, LLM, HTTP-обвязка) покрыта на 84–94%. Пакеты, работающие с реальной БД (`repository`, `database`, основной маршрутизатор `cmd/api`), сознательно покрываются `make test-e2e` через testcontainers — там запускается весь HTTP-стек поверх настоящего Postgres, поэтому unit-coverage по этим пакетам остаётся низким, но фактическое поведение проверено.

### Frontend

| Команда | Что делает |
| --- | --- |
| `pnpm -C frontend install` | Установка зависимостей (требует `pnpm install --frozen-lockfile` в CI) |
| `pnpm -C frontend lint` | ESLint |
| `pnpm -C frontend typecheck` | `tsc -b --pretty false` |
| `pnpm -C frontend test` | Vitest (jsdom + Testing Library) |
| `pnpm -C frontend build` | Production-сборка через Vite |

Coverage:

```bash
cd frontend
pnpm exec vitest run --coverage
```

Текущее состояние: **6 test files, 23 tests passed**. Тесты сосредоточены на пограничных и легко-тестируемых модулях:

- `src/components/ProtectedRoute.test.tsx` — гард авторизации и ролей.
- `src/components/RunStatusBadge.test.tsx` — маппинг статусов запусков.
- `src/lib/forms/schemas/assistantSchema.test.ts` — Zod-схема формы ассистента.
- `src/lib/forms/schemas/authSchema.test.ts` — Zod-схема формы входа и регистрации.
- `src/lib/format.test.ts` — форматтеры дат/значений.
- `src/lib/hooks/useDebouncedValue.test.ts` — дебаунс-хук.

Frontend unit-тесты сфокусированы на критичных и стабильных частях приложения:схемах валидации, форматтерах, хуках и небольших UI-компонентах. Крупные UI-страницы и большинство shadcn/ui-компонентов не покрываются unit-тестами напрямую и проверяются интеграционно во время ручного тестирования приложения и backend e2e-сценариев.

### Итоговое покрытие

| Часть проекта | Coverage |
| --- | --- |
| Backend (`go test -covermode=atomic ./...`) | **60.3%** |
| Frontend (`vitest --coverage`) | **2.74% line coverage** |

Backend покрывает основную доменную и HTTP-логику unit-тестами, а интеграционные сценарии дополнительно проверяются через e2e-тесты с реальным PostgreSQL и полным HTTP-стеком. Frontend unit-тесты сфокусированы на guards, схемах валидации, форматтерах и вспомогательных хуках.

### CI

GitHub Actions workflow `.github/workflows/ci.yml` повторяет локальные команды:

- Job **Backend**: `make generate` + проверка `git diff` для OpenAPI Go-типов → `make build` → `golangci-lint` → `go test -race -coverprofile` → `make test-e2e` (testcontainers поднимает Postgres в Docker) → артефакт `backend-coverage`.
- Job **Frontend**: `pnpm install --frozen-lockfile` → `pnpm gen:api` + проверка `git diff` для OpenAPI TypeScript-типов → `pnpm lint` → `pnpm typecheck` → `pnpm exec vitest run --coverage` → артефакт `frontend-coverage` → `pnpm build`.

Запускается на push и PR в `main`/`master`. Concurrency-группа отменяет устаревшие раны для той же ветки.

Проверка отклонения схемы OpenAPI — отдельное решение: `api.yaml` остаётся источником истины, а CI падает, если сгенерированные backend/frontend типы не обновлены после изменения контракта.

## Нагрузочное тестирование

Сценарий k6 для `GET /assistants` лежит в `load-tests/assistants_list.js`. Тестирует «горячий» endpoint каталога с разными комбинациями параметров: базовая пагинация, увеличенный `pageSize`, фильтр `categoryId`, поиск `q` (использует `pg_trgm`-индекс).

### Запуск

Backend должен быть поднят (`docker compose up --build`), категории и ассистенты — засеяны.

```bash
k6 run load-tests/assistants_list.js
# с другим хостом или ролью:
BASE_URL=http://localhost:8080 ROLE=admin k6 run load-tests/assistants_list.js
```

### Профиль нагрузки

Ramping-VUs: 0 → 20 за 15 с → 50 за 30 с → 0 за 15 с (всего ~60 с). Каждая итерация VU делает 3–4 разнотипных запроса к `/assistants`.

### Пороги (`thresholds`)

- `http_req_failed < 1%`
- p95 latency `/assistants` < 300 ms
- p99 latency `/assistants` < 800 ms
- доля успешных `check` > 99%

Если любой порог не выполняется, k6 завершает run с ненулевым кодом — это удобно встраивать в CI/PR-проверки.

### Результаты прогона

Локально (Docker compose на macOS, M-серия, Postgres в контейнере):

| Метрика | Значение |
| --- | --- |
| Всего HTTP-запросов | 13 594 за 60 с |
| RPS | ≈ 226 |
| Ошибок | 0 (0.00%) |
| Latency `/assistants` avg | 1.51 ms |
| Latency `/assistants` p90 | 2.43 ms |
| Latency `/assistants` p95 | 3.02 ms |
| Latency `/assistants` p99 | 8.63 ms |
| Latency `/assistants` max | 29.21 ms |
| Доля успешных `check` | 100% |
