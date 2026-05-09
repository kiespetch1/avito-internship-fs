CREATE TABLE assistants (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id         UUID        NOT NULL REFERENCES categories(id),
    name                TEXT        NOT NULL,
    description         TEXT        NOT NULL DEFAULT '',
    model               TEXT        NOT NULL,
    system_prompt       TEXT        NOT NULL,
    example_user_prompt TEXT,
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Фильтрация по категории
CREATE INDEX idx_assistants_category_id ON assistants (category_id);

-- Индекс для запроса активных
CREATE INDEX idx_assistants_active ON assistants (is_active) WHERE is_active = TRUE;

-- Триграммные индексы для подстрочного поиска (ILIKE '%запрос%')
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_assistants_name_trgm ON assistants USING GIN (name gin_trgm_ops);
CREATE INDEX idx_assistants_desc_trgm ON assistants USING GIN (description gin_trgm_ops);
