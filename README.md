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

## Решения по API

### `systemPrompt` скрыт от обычного пользователя

В контракте `Assistant.systemPrompt` помечен `nullable: true`. Backend возвращает реальное значение только для роли `admin`, для роли `user` поле сериализуется как `null`.

Системный промпт — внутренняя настройка ассистента, в большинстве случаев пользователю не стоит ее знать как минимум из-за находящихся там формулировок, ограничений, правил, в том числе ограничивающих обходы). В БД поле хранится как `NOT NULL` — на стороне сервиса оно всегда есть.

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
