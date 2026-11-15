package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bytebytego-learning/design-a-url-shortner/internal/shortcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	store := &Postgres{pool: pool}
	if err := store.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := store.initialize(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) initialize(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    normalized_url TEXT NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('full', 'path_only')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (normalized_url, mode)
)`)
	return err
}

func (p *Postgres) FindByKey(ctx context.Context, normalized, mode string) (Record, bool, error) {
	return p.find(ctx, `SELECT id, short_code, original_url, normalized_url, mode FROM urls WHERE normalized_url=$1 AND mode=$2`, normalized, mode)
}

func (p *Postgres) FindByCode(ctx context.Context, code string) (Record, bool, error) {
	return p.find(ctx, `SELECT id, short_code, original_url, normalized_url, mode FROM urls WHERE short_code=$1`, code)
}

func (p *Postgres) find(ctx context.Context, query string, args ...any) (Record, bool, error) {
	var record Record
	err := p.pool.QueryRow(ctx, query, args...).Scan(&record.ID, &record.ShortCode, &record.OriginalURL, &record.NormalizedURL, &record.Mode)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	return record, true, nil
}

func (p *Postgres) CreateOrGet(ctx context.Context, record Record) (Record, bool, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return Record{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	if err := tx.QueryRow(ctx, "SELECT nextval('urls_id_seq')").Scan(&id); err != nil {
		return Record{}, false, err
	}
	code, err := shortcode.Encode(id)
	if err != nil {
		return Record{}, false, err
	}
	result, err := tx.Exec(ctx, `
INSERT INTO urls (id, short_code, original_url, normalized_url, mode)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (normalized_url, mode) DO NOTHING`, id, code, record.OriginalURL, record.NormalizedURL, record.Mode)
	if err != nil {
		return Record{}, false, err
	}
	if result.RowsAffected() == 0 {
		existing, found, findErr := findTx(ctx, tx, record.NormalizedURL, record.Mode)
		if findErr != nil || !found {
			return Record{}, false, fmt.Errorf("find existing URL after conflict: %w", findErr)
		}
		if err := tx.Commit(ctx); err != nil {
			return Record{}, false, err
		}
		return existing, true, nil
	}
	if err := tx.Commit(ctx); err != nil {
		return Record{}, false, err
	}
	record.ID = id
	record.ShortCode = code
	return record, false, nil
}

func findTx(ctx context.Context, tx pgx.Tx, normalized, mode string) (Record, bool, error) {
	var record Record
	err := tx.QueryRow(ctx, `SELECT id, short_code, original_url, normalized_url, mode FROM urls WHERE normalized_url=$1 AND mode=$2`, normalized, mode).Scan(&record.ID, &record.ShortCode, &record.OriginalURL, &record.NormalizedURL, &record.Mode)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, false, nil
	}
	return record, err == nil, err
}
