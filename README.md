# Каталог AI-ассистентов

Fullstack-приложение для управления, просмотра и запуска AI-ассистентов.

## Запуск

```bash
docker compose up --build
```

- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- Healthcheck: `GET http://localhost:8080/_info`

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
  timeout: 60s
  defaultModel: gpt-mock
  baseUrl: ""
  apiKey: ""
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
| `LLM_PROVIDER`       | `mock`                                                                        | Идентификатор LLM-провайдера                     |
| `LLM_TIMEOUT`        | `60s`                                                                         | Таймаут запроса к LLM                            |
| `LLM_DEFAULT_MODEL`  | `gpt-mock`                                                                    | Модель по умолчанию для mock                     |
| `LLM_BASE_URL`       | —                                                                             | URL OpenAI-совместимого провайдера (опционально) |
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

### Обработка запуска ассистента и сохранение истории

Реализовано в `backend/internal/service/runs.go`. Три независимых контекста с таймаутами:

1. **Проверка ассистента** — в рамках HTTP-запроса.
2. **Сохранение `pending`** — `context.Background()` + 5 с. Запись происходит до вызова LLM; если БД недоступна — HTTP 500 без утечки.
3. **Вызов LLM** — `context.Background()` + `LLM_TIMEOUT` (60 с по умолчанию). Контекст **не** привязан к HTTP-запросу: если клиент отключится, LLM-вызов продолжится.
4. **Финализация** — ещё один `context.Background()` + 5 с. Запуск переводится в `success` или `failed` независимо от жизни соединения.

Такая схема гарантирует, что запись в `assistant_runs` никогда не зависает в `pending` из-за разрыва клиентского соединения.

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

type Provider interface {
    Generate(ctx context.Context, req Request) (Response, error)
}
```

`MockProvider` (`llm/mock.go`) реализует интерфейс без сетевых вызовов:
- возвращает `[mock:<model>] <userPrompt>` как `Output`;
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

### Переключение с мок-провайдера на внешний OpenAI-совместимый API

Логика выбора провайдера сосредоточена в `resolveLLMProvider` (`cmd/api/main.go`). Чтобы добавить поддержку OpenAI-совместимого провайдера:

1. Реализовать `llm.Provider` для HTTP-клиента
2. Добавить ветку в `resolveLLMProvider`:

```go
case "openai":
    return llm.NewOpenAIProvider(cfg.BaseURL, cfg.APIKey, cfg.DefaultModel), nil
```

3. Переключить провайдер через переменные окружения — без пересборки образа:

```bash
LLM_PROVIDER=openai
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-...
LLM_DEFAULT_MODEL=gpt-4o
```

`LLM_BASE_URL` позволяет указать любой OpenAI-совместимый эндпоинт.

### Стейт-менеджмент на frontend

Вся серверная асинхронная логика управляется **TanStack Query**:
- данные кешируются на 30 с (`staleTime: 30_000`), `refetchOnWindowFocus` отключён;
- повторные попытки не выполняются для 401 / 403 / 404, для остальных ошибок — до 2 повторов;
- глобальный обработчик в `QueryCache` / `MutationCache` разлогинивает пользователя при 401.

Аутентификационное состояние (JWT-токен + профиль пользователя) живёт в `localStorage` под ключом `avito.auth.v1` и читается через **`useSyncExternalStore`** (без дополнительных библиотек, поскольку приложение не предполагает сложного стейта). Изменения распространяются на все открытые вкладки через `window.addEventListener('storage', ...)`. Хранилище — `frontend/src/lib/auth/storage.ts`.

Формы управляются через **TanStack Form** + **Zod** (валидация схем в `frontend/src/lib/forms/schemas/`).

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
