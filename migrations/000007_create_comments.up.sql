CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users (id),
    dot_id      UUID        REFERENCES dots (id) ON DELETE CASCADE,
    object_id   UUID        REFERENCES objects (id) ON DELETE CASCADE,
    path_id     UUID        REFERENCES paths (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    body        VARCHAR(240) NOT NULL DEFAULT '',
    -- async post-processing (email + audit) flips this to completed/failed
    status      VARCHAR(20) NOT NULL DEFAULT 'processing'
        CHECK (status IN ('processing', 'completed', 'failed')),

    -- a comment belongs to exactly one entity
    CONSTRAINT comments_exactly_one_target CHECK (num_nonnulls(dot_id, object_id, path_id) = 1)
);

-- FK columns are not auto-indexed; every comment lookup filters by one of these
CREATE INDEX comments_dot_id_idx ON comments (dot_id);
CREATE INDEX comments_object_id_idx ON comments (object_id);
CREATE INDEX comments_path_id_idx ON comments (path_id);

INSERT INTO comments (id, user_id, object_id, body)
VALUES (
    'e3b6d9a2-1c4f-4e8a-9b7d-5f2a0c8e6d41',
    '9f27aa45-efd3-40e7-bba1-72cc9004de5f',
    '03c7faf2-0000-0000-0000-000000000000',
    'Пример комментария'
);
