// Package postgres implements store.Store on top of PostgreSQL using the
// database/sql standard library and the lib/pq driver. It is the durable
// backend intended for multi-node deployments where at-least-once delivery
// must survive process restarts.
package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
)

//go:embed schema.sql
var schema string

// Store is a PostgreSQL-backed store.Store.
type Store struct {
	db *sql.DB
}

var _ store.Store = (*Store)(nil)

// Open connects to the database at dsn, verifies connectivity and applies the
// schema. The caller owns the returned Store and must Close it.
func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying connection pool.
func (s *Store) Close() error { return s.db.Close() }

// PutJob inserts or replaces a job.
func (s *Store) PutJob(ctx context.Context, j job.Job) error {
	command, _ := json.Marshal(nonNil(j.Command))
	backoff, _ := json.Marshal(j.Backoff)
	deps, _ := json.Marshal(nonNil(j.DependsOn))
	created := orNow(j.CreatedAt)
	updated := orNow(j.UpdatedAt)

	_, err := s.db.ExecContext(ctx, `
INSERT INTO jobs (id, name, spec, command, max_retries, backoff, depends_on, timeout, enabled, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    spec = EXCLUDED.spec,
    command = EXCLUDED.command,
    max_retries = EXCLUDED.max_retries,
    backoff = EXCLUDED.backoff,
    depends_on = EXCLUDED.depends_on,
    timeout = EXCLUDED.timeout,
    enabled = EXCLUDED.enabled,
    updated_at = EXCLUDED.updated_at`,
		j.ID, j.Name, j.Spec, command, j.MaxRetries, backoff, deps,
		int64(j.Timeout), j.Enabled, created, updated)
	return err
}

// GetJob returns a job by ID.
func (s *Store) GetJob(ctx context.Context, id string) (job.Job, error) {
	row := s.db.QueryRowContext(ctx, jobColumns+` FROM jobs WHERE id = $1`, id)
	j, err := scanJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return job.Job{}, store.ErrNotFound
	}
	return j, err
}

// ListJobs returns every job ordered by ID.
func (s *Store) ListJobs(ctx context.Context) ([]job.Job, error) {
	rows, err := s.db.QueryContext(ctx, jobColumns+` FROM jobs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []job.Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// DeleteJob removes a job by ID.
func (s *Store) DeleteJob(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return store.ErrNotFound
	}
	return nil
}

const jobColumns = `SELECT id, name, spec, command, max_retries, backoff, depends_on, timeout, enabled, created_at, updated_at`

func scanJob(sc scanner) (job.Job, error) {
	var (
		j                      job.Job
		command, backoff, deps []byte
		timeoutNs              int64
	)
	if err := sc.Scan(&j.ID, &j.Name, &j.Spec, &command, &j.MaxRetries, &backoff,
		&deps, &timeoutNs, &j.Enabled, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return job.Job{}, err
	}
	j.Timeout = time.Duration(timeoutNs)
	_ = json.Unmarshal(command, &j.Command)
	_ = json.Unmarshal(backoff, &j.Backoff)
	_ = json.Unmarshal(deps, &j.DependsOn)
	return j, nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func orNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

const runColumns = `SELECT id, job_id, batch_id, state, attempt, scheduled_for, started_at, finished_at, lease_expiry, worker, error, created_at, updated_at`

// CreateRun stores a new run.
func (s *Store) CreateRun(ctx context.Context, r job.Run) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO runs (id, job_id, batch_id, state, attempt, scheduled_for, started_at, finished_at, lease_expiry, worker, error, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		r.ID, r.JobID, r.BatchID, r.State, r.Attempt, r.ScheduledFor,
		r.StartedAt, r.FinishedAt, nullTime(r.LeaseExpiry), r.Worker, r.Error,
		orNow(r.CreatedAt), orNow(r.UpdatedAt))
	return err
}

// GetRun returns a run by ID.
func (s *Store) GetRun(ctx context.Context, id string) (job.Run, error) {
	row := s.db.QueryRowContext(ctx, runColumns+` FROM runs WHERE id = $1`, id)
	r, err := scanRun(row)
	if errors.Is(err, sql.ErrNoRows) {
		return job.Run{}, store.ErrNotFound
	}
	return r, err
}

// UpdateRun persists changes to an existing run.
func (s *Store) UpdateRun(ctx context.Context, r job.Run) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE runs SET job_id=$2, batch_id=$3, state=$4, attempt=$5, scheduled_for=$6,
    started_at=$7, finished_at=$8, lease_expiry=$9, worker=$10, error=$11, updated_at=$12
WHERE id=$1`,
		r.ID, r.JobID, r.BatchID, r.State, r.Attempt, r.ScheduledFor,
		r.StartedAt, r.FinishedAt, nullTime(r.LeaseExpiry), r.Worker, r.Error,
		orNow(r.UpdatedAt))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return store.ErrNotFound
	}
	return nil
}

// ListRuns returns runs matching the filter, newest first.
func (s *Store) ListRuns(ctx context.Context, f store.RunFilter) ([]job.Run, error) {
	query := runColumns + ` FROM runs`
	var conds []string
	var args []any
	add := func(clause string, val any) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(clause, len(args)))
	}
	if f.JobID != "" {
		add("job_id = $%d", f.JobID)
	}
	if f.BatchID != "" {
		add("batch_id = $%d", f.BatchID)
	}
	if f.State != "" {
		add("state = $%d", string(f.State))
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if f.Offset > 0 {
		args = append(args, f.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]job.Run, 0)
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ClaimDue leases up to limit ready runs and marks them running. It uses
// FOR UPDATE SKIP LOCKED so that many workers can claim disjoint sets of runs
// concurrently without blocking one another, which is what makes the
// PostgreSQL backend safe to run on multiple nodes.
func (s *Store) ClaimDue(ctx context.Context, now time.Time, worker string, lease time.Duration, limit int) ([]job.Run, error) {
	rows, err := s.db.QueryContext(ctx, `
WITH ready AS (
    SELECT id FROM runs
    WHERE (state = 'pending' AND scheduled_for <= $1)
       OR (state = 'running' AND lease_expiry < $1)
    ORDER BY scheduled_for
    LIMIT $4
    FOR UPDATE SKIP LOCKED
)
UPDATE runs r
SET state = 'running', worker = $2, lease_expiry = $3, updated_at = $1,
    started_at = COALESCE(r.started_at, $1)
FROM ready
WHERE r.id = ready.id
RETURNING r.id, r.job_id, r.batch_id, r.state, r.attempt, r.scheduled_for,
          r.started_at, r.finished_at, r.lease_expiry, r.worker, r.error,
          r.created_at, r.updated_at`,
		now, worker, now.Add(lease), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []job.Run
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanRun(sc scanner) (job.Run, error) {
	var (
		r                        job.Run
		started, finished, lease sql.NullTime
	)
	if err := sc.Scan(&r.ID, &r.JobID, &r.BatchID, &r.State, &r.Attempt, &r.ScheduledFor,
		&started, &finished, &lease, &r.Worker, &r.Error, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return job.Run{}, err
	}
	if started.Valid {
		t := started.Time
		r.StartedAt = &t
	}
	if finished.Valid {
		t := finished.Time
		r.FinishedAt = &t
	}
	if lease.Valid {
		r.LeaseExpiry = lease.Time
	}
	return r, nil
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
