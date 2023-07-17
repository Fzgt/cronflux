-- cronflux PostgreSQL schema. Applied idempotently on startup.

CREATE TABLE IF NOT EXISTS jobs (
    id          text PRIMARY KEY,
    name        text        NOT NULL DEFAULT '',
    spec        text        NOT NULL DEFAULT '',
    command     jsonb       NOT NULL DEFAULT '[]',
    max_retries int         NOT NULL DEFAULT 0,
    backoff     jsonb       NOT NULL DEFAULT '{}',
    depends_on  jsonb       NOT NULL DEFAULT '[]',
    timeout     bigint      NOT NULL DEFAULT 0,
    enabled     boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS runs (
    id            text PRIMARY KEY,
    job_id        text        NOT NULL,
    batch_id      text        NOT NULL DEFAULT '',
    state         text        NOT NULL,
    attempt       int         NOT NULL DEFAULT 0,
    scheduled_for timestamptz NOT NULL,
    started_at    timestamptz,
    finished_at   timestamptz,
    lease_expiry  timestamptz,
    worker        text        NOT NULL DEFAULT '',
    error         text        NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS runs_claim_idx ON runs (state, scheduled_for);
CREATE INDEX IF NOT EXISTS runs_job_idx ON runs (job_id);
CREATE INDEX IF NOT EXISTS runs_batch_idx ON runs (batch_id);
