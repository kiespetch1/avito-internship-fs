CREATE TABLE run_feedback (
    id         UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id     UUID     NOT NULL UNIQUE REFERENCES assistant_runs(id) ON DELETE CASCADE,
    rating     SMALLINT NOT NULL CHECK (rating IN (-1, 1)),  -- -1 дизлайк, 1 лайк
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
