package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(databaseURL string) (*DB, error) {
	ctx := context.Background()

	// Configure connection pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing database URL: %w", err)
	}

	// Set pool config
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connection pool established")

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}