package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Storage struct {
	writeDB *sql.DB
	readDB  *sql.DB
}

const (
	maxRetries    = 3
	retryInterval = 500 * time.Millisecond
)

func NewStorage(writeDB, readDB *sql.DB) *Storage {
	return &Storage{writeDB: writeDB, readDB: readDB}
}

func (s *Storage) Close() error {
	if err := s.writeDB.Close(); err != nil {
		return err
	}
	return s.readDB.Close()
}

func (s *Storage) HealthCheck(ctx context.Context) error {
	if err := s.writeDB.PingContext(ctx); err != nil {
		return fmt.Errorf("writeDB ping failed: %w", err)
	}
	if err := s.readDB.PingContext(ctx); err != nil {
		return fmt.Errorf("readDB ping failed: %w", err)
	}
	return nil
}

func withRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryInterval * time.Duration(attempt)):
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("all %d attempts failed, last error: %w", maxRetries, lastErr)
}
