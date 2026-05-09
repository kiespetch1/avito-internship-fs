CREATE TABLE assistant_runs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    assistant_id UUID        NOT NULL REFERENCES assistants(id),
    user_id      UUID        NOT NULL REFERENCES users(id),
    model        TEXT        NOT NULL,
    user_prompt  TEXT        NOT NULL,
    output       TEXT,
    status       TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed')),
    error        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Индекс по истории запусков пользователя, отсортированной по дате
CREATE INDEX idx_runs_user_created ON assistant_runs (user_id, created_at DESC);

-- Индекс для фильтрации по статусу
CREATE INDEX idx_runs_status ON assistant_runs (status);

-- Индекс для фильтрации по ассистенту (админская панель).
CREATE INDEX idx_runs_assistant ON assistant_runs (assistant_id);
