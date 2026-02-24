package db

import (
	"context"
	"database/sql"
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

type DBHealth struct {
	MasterUp        bool `json:"masterUp"`
	ReplicaUp       bool `json:"replicaUp"`
	MasterRecovery  bool `json:"masterInRecovery"`
	ReplicaRecovery bool `json:"replicaInRecovery"`
}

func NewStorage(writeDB, readDB *sql.DB) *Storage {
	return &Storage{writeDB: writeDB, readDB: readDB}
}

func (s *Storage) Close() error {
	if err := s.writeDB.Close(); err != nil {
		return err
	}
	return s.readDB.Close()
}

func (s *Storage) HealthCheck(ctx context.Context) (*DBHealth, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	health := &DBHealth{}

	if err := s.writeDB.PingContext(ctx); err == nil {
		health.MasterUp = true

		var inRecovery bool
		s.writeDB.QueryRowContext(ctx,
			"SELECT pg_is_in_recovery()").
			Scan(&inRecovery)

		health.MasterRecovery = inRecovery
	}

	if err := s.readDB.PingContext(ctx); err == nil {
		health.ReplicaUp = true

		var inRecovery bool
		s.readDB.QueryRowContext(ctx,
			"SELECT pg_is_in_recovery()").
			Scan(&inRecovery)

		health.ReplicaRecovery = inRecovery
	}

	return health, nil
}

func (s *Storage) retryRead(
	ctx context.Context,
	fn func(db *sql.DB) error,
) error {

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {

		err := fn(s.readDB)
		if err == nil {
			return nil
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval * time.Duration(attempt)):
		}
	}

	return lastErr
}

func (s *Storage) readWithFallback(
	ctx context.Context,
	fn func(db *sql.DB) error,
) error {

	err := s.retryRead(ctx, fn)
	if err == nil {
		return nil
	}

	return fn(s.writeDB)
}
