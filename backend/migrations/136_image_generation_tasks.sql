-- Image Studio async task queue, cached assets, and upstream diagnostics.

CREATE TABLE IF NOT EXISTS image_generation_tasks (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    group_id BIGINT NULL REFERENCES groups(id) ON DELETE SET NULL,
    mode VARCHAR(32) NOT NULL,
    model VARCHAR(100) NOT NULL,
    prompt TEXT NOT NULL,
    ratio VARCHAR(16) NOT NULL,
    resolution VARCHAR(8) NOT NULL,
    size VARCHAR(32) NOT NULL,
    quality VARCHAR(32) NOT NULL DEFAULT 'high',
    count INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    progress INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NULL,
    request_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    account_id BIGINT NULL REFERENCES accounts(id) ON DELETE SET NULL,
    usage_meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT image_generation_tasks_count_check CHECK (count >= 1 AND count <= 4),
    CONSTRAINT image_generation_tasks_progress_check CHECK (progress >= 0 AND progress <= 100)
);

CREATE INDEX IF NOT EXISTS idx_image_generation_tasks_user_created
    ON image_generation_tasks(user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_image_generation_tasks_status_created
    ON image_generation_tasks(status, created_at)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS image_generation_assets (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(64) NOT NULL REFERENCES image_generation_tasks(task_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seq INTEGER NOT NULL DEFAULT 0,
    kind VARCHAR(20) NOT NULL DEFAULT 'output',
    storage_driver VARCHAR(20) NOT NULL,
    storage_key TEXT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    original_url TEXT NULL,
    revised_prompt TEXT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_generation_assets_task_seq
    ON image_generation_assets(task_id, kind, seq)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_image_generation_assets_user
    ON image_generation_assets(user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS image_generation_upstream_logs (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(64) NOT NULL REFERENCES image_generation_tasks(task_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT NULL REFERENCES accounts(id) ON DELETE SET NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'openai',
    endpoint VARCHAR(128) NOT NULL,
    status_code INTEGER NULL,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    request_excerpt TEXT NOT NULL DEFAULT '',
    response_excerpt TEXT NOT NULL DEFAULT '',
    error_message TEXT NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_generation_upstream_logs_task
    ON image_generation_upstream_logs(task_id, created_at DESC);

