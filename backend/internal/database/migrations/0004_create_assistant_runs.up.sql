CREATE TABLE assistant_runs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    assistant_id  UUID        NOT NULL REFERENCES assistants(id),
    user_id       UUID        NOT NULL REFERENCES users(id),
    model         TEXT        NOT NULL,
    user_prompt   TEXT        NOT NULL,
    output        TEXT,
    status        TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed')),
    error         TEXT,
    tokens_in     INTEGER,
    tokens_out    INTEGER,
    latency_ms    INTEGER,
    finish_reason TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- История запусков пользователя, отсортированная по дате
CREATE INDEX idx_runs_user_created ON assistant_runs (user_id, created_at DESC);

-- Фильтрация по статусу
CREATE INDEX idx_runs_status ON assistant_runs (status);

-- Фильтрация по ассистенту (админская панель)
CREATE INDEX idx_runs_assistant ON assistant_runs (assistant_id);