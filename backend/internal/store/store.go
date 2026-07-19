package store

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, cfg Config) (*Store, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.dsn())
	if err != nil {
		return nil, fmt.Errorf("store: parse config: %w", err)
	}

	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}
	timeoutCtx, cancelCtx := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCtx()
	err = pool.Ping(timeoutCtx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}
