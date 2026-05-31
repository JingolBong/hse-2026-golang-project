package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"

	applog "hse-2026-golang-project/internal/logger"
)

func main() {
	logger := applog.New(applog.Options{
		Level:   os.Getenv("LOG_LEVEL"),
		Service: "migrator",
		ToFile:  applog.ToFileEnabled(),
	})

	logger.Info("Starting migrator...")

	dsn := "host=postgres-master port=5432 user=postgres password=postgres dbname=testdb sslmode=disable"

	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		logger.WithField("attempt", i+1).Warn("Waiting for DB to be ready...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Connected to DB, running migrations...")

	files, err := os.ReadDir("migrations")
	if err != nil {
		logger.Fatalf("Failed to read migrations directory: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		logger.WithField("file", file).Info("Executing migration")
		content, err := os.ReadFile(filepath.Join("migrations", file))
		if err != nil {
			logger.Fatalf("Failed to read file %s: %v", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			logger.Fatalf("Migration %s failed: %v", file, err)
		}
	}

	logger.Info("All migrations applied successfully!")
}
