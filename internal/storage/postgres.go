package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct{ db *sql.DB }

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil { return nil, err }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil { return nil, err }
	p := &Postgres{db: db}
	if err := p.migrate(ctx); err != nil { return nil, err }
	return p, nil
}

func (p *Postgres) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS secrets ( path TEXT PRIMARY KEY )`,
		`CREATE TABLE IF NOT EXISTS versions (
			id BIGSERIAL PRIMARY KEY,
			version_id TEXT NOT NULL,
			path TEXT NOT NULL REFERENCES secrets(path) ON DELETE CASCADE,
			ciphertext BYTEA NOT NULL,
			wrapped_dek BYTEA NOT NULL,
			kek_id TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS versions_path_created_idx ON versions(path, created_at DESC)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS versions_path_version_idx ON versions(path, version_id)`,
	}
	for _, s := range stmts { if _, err := p.db.ExecContext(ctx, s); err != nil { return err } }
	return nil
}

func (p *Postgres) Put(path string, v SecretVersion) (string, error) {
	if path == "" { return "", errors.New("empty path") }
	if v.VersionID == "" { v.VersionID = genVersionID() }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := p.db.ExecContext(ctx, `INSERT INTO secrets(path) VALUES ($1) ON CONFLICT DO NOTHING`, path); err != nil {
		return "", err
	}
	_, err := p.db.ExecContext(ctx, `INSERT INTO versions(version_id, path, ciphertext, wrapped_dek, kek_id, created_at) VALUES ($1,$2,$3,$4,$5, to_timestamp($6))`, v.VersionID, path, v.Ciphertext, v.WrappedDEK, v.KEKID, v.CreatedAtIfZeroNow())
	if err != nil { return "", err }
	return v.VersionID, nil
}

func (p *Postgres) Get(path, version string) (*SecretVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := `SELECT version_id, ciphertext, wrapped_dek, kek_id, EXTRACT(EPOCH FROM created_at)::bigint FROM versions WHERE path=$1`
	args := []any{path}
	if version == "" || version == "latest" {
		q += ` ORDER BY created_at DESC, id DESC LIMIT 1`
	} else {
		q += ` AND version_id=$2`
		args = append(args, version)
	}
	var out SecretVersion
	if err := p.db.QueryRowContext(ctx, q, args...).Scan(&out.VersionID, &out.Ciphertext, &out.WrappedDEK, &out.KEKID, &out.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, fmt.Errorf("not found") }
		return nil, err
	}
	return &out, nil
}

func (v *SecretVersion) CreatedAtIfZeroNow() int64 { if v.CreatedAt == 0 { v.CreatedAt = time.Now().Unix() }; return v.CreatedAt }

func genVersionID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
