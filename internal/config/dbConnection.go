package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(ctx context.Context, dbCfg DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Database, dbCfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	var pingErr error
	for i := 0; i < 5; i++ {
		pingErr = db.PingContext(ctx)
		if pingErr == nil {
			return db, nil
		}

		log.Printf("DB (%s) not ready: %v. Retrying... (%d/5)", dbCfg.Host, pingErr, i+1)
		select {
		case <-ctx.Done():
			db.Close()
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}

	db.Close()
	return nil, fmt.Errorf("db connection failed after retries: %w", pingErr)
}
