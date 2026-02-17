package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(ctx context.Context, cfg *Config) (*sql.DB, error) {

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBSettings.User,
		cfg.DBSettings.Password,
		cfg.DBSettings.Host,
		cfg.DBSettings.Port,
		cfg.DBSettings.Database,
	)

	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("sql.Open: %w", err)
		}

		err = db.PingContext(ctx)
		if err == nil {
			break
		}

		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("db connection failed after retries: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
