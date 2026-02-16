package db

import (
	"database/sql"
	"fmt"
)

func (s *Storage) UpsertProject(p Project) (int, error) {
	const insertQuery = `
        INSERT INTO project (key, name, url)
        VALUES ($1, $2, $3)
        ON CONFLICT (key) DO NOTHING
        RETURNING id;
    `

	var id int

	err := s.db.QueryRow(insertQuery, p.Key, p.Name, p.URL).Scan(&id)
	if err == nil {
		return id, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("insert project %s: %w", p.Key, err)
	}

	const selectQuery = `
        SELECT id FROM project
        WHERE key = $1;
    `
	err = s.db.QueryRow(selectQuery, p.Key).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("select project %s: %w", p.Key, err)
	}

	return id, nil
}
func (s *Storage) UpsertAuthor(a Author) (int, error)                    {}
func (s *Storage) BatchInsertIssues(projectID int, issues []Issue) error {}
func (s *Storage) BatchInsertStatusChanges(changes []StatusChange) error {}
func (s *Storage) GetProjectByKey(key string) (*Project, error)          {}
func (s *Storage) GetIssuesByProject(projectID int) ([]Issue, error)     {}
