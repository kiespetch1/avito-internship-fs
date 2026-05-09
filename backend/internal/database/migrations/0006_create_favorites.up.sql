CREATE TABLE favorite_assistants (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assistant_id UUID NOT NULL REFERENCES assistants(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, assistant_id)
);
