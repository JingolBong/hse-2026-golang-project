package db

import (
	"database/sql"
)

type Storage struct {
	db *sql.DB
}
