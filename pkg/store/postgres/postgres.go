package postgres

import (
	"context"
	"time"

	"gogogo/config"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

type Store struct {
	Pool *pgxpool.Pool
}

func New(c *config.Config) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(c.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = conn.Ping(ctx); err != nil {
		return nil, err
	}

	return &Store{Pool: conn}, nil
}
