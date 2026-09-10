CREATE TABLE IF NOT EXISTS usage_stats (
    id            INTEGER PRIMARY KEY DEFAULT 1,
    total_runs    BIGINT NOT NULL DEFAULT 0,
    total_elements BIGINT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

INSERT INTO usage_stats (id) VALUES (1) ON CONFLICT DO NOTHING;
