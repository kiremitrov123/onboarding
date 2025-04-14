package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Postgres struct {
	bun *bun.DB
}

// Returns the underlying *bun.DB instance.
func (p *Postgres) DB() *bun.DB {
	return p.bun
}

// NewPostgres connects to the database via pgdriver
// pings to ensure the connection is up and running
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	db, err := waitForDB(ctx, dsn, 10, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Postgres{
		bun: db,
	}, nil
}

// Retries connecting to the database with delay.
func waitForDB(ctx context.Context, dsn string, retries int, delay time.Duration) (*bun.DB, error) {
	for i := 0; i < retries; i++ {
		sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

		if err := sqlDB.PingContext(ctx); err == nil {
			log.Printf("Connected to CockroachDB on attempt %d", i+1)
			return bun.NewDB(sqlDB, pgdialect.New()), nil
		} else {
			log.Printf("⏳ DB not ready yet (attempt %d/%d): %v", i+1, retries, err)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// retry
		}
	}

	return nil, fmt.Errorf("DB not reachable after %d retries", retries)
}
