CREATE TABLE tags (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assistant_tags (
    assistant_id UUID NOT NULL REFERENCES assistants(id) ON DELETE CASCADE,
    tag_id       UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (assistant_id, tag_id)
);

-- Индекс на поиск ассистентов по тегу
CREATE INDEX idx_assistant_tags_tag ON assistant_tags (tag_id);
