package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ProjectID       uuid.UUID  `json:"project_id"`
	UserID          uuid.UUID  `json:"user_id"`
	ProjectName     string     `json:"project_name"`
	Color           *string    `json:"color"`
	IsInbox         *bool      `json:"is_inbox"`
	ParentProjectID *uuid.UUID `json:"parent_project_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ProjectModel struct {
	DB *pgxpool.Pool
}

func (m *ProjectModel) AddProject(project Project) (Project, error) {
	query := `
		INSERT INTO projects (
			user_id, project_name, color, is_inbox, parent_project_id
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING project_id, user_id, project_name, color, is_inbox, parent_project_id, created_at
	`

	var createdProject Project
	err := m.DB.QueryRow(
		context.Background(),
		query,
		project.UserID,
		project.ProjectName,
		project.Color,
		project.IsInbox,
		project.ParentProjectID,
	).Scan(
		&createdProject.ProjectID,
		&createdProject.UserID,
		&createdProject.ProjectName,
		&createdProject.Color,
		&createdProject.IsInbox,
		&createdProject.ParentProjectID,
		&createdProject.CreatedAt,
	)

	if err != nil {
		return Project{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return createdProject, nil
}

func (m *ProjectModel) EditProjectByID(project Project) (Project, error) {
	query := `
		UPDATE projects SET
			project_name = $3,
			color = $4,
			is_inbox = $5,
			parent_project_id = $6
		WHERE project_id = $1 AND user_id = $2
		RETURNING project_id, user_id, project_name, color, is_inbox, parent_project_id, created_at
	`

	var updatedProject Project
	err := m.DB.QueryRow(
		context.Background(),
		query,
		project.ProjectID,
		project.UserID,
		project.ProjectName,
		project.Color,
		project.IsInbox,
		project.ParentProjectID,
	).Scan(
		&updatedProject.ProjectID,
		&updatedProject.UserID,
		&updatedProject.ProjectName,
		&updatedProject.Color,
		&updatedProject.IsInbox,
		&updatedProject.ParentProjectID,
		&updatedProject.CreatedAt,
	)

	if err != nil {
		return Project{}, fmt.Errorf("unable to execute query: %v", err)
	}

	return updatedProject, nil
}

func (m *ProjectModel) GetProjectsByUserID(userID uuid.UUID) ([]Project, error) {
	query := `
		SELECT project_id, user_id, project_name, color, is_inbox, parent_project_id, created_at
		FROM projects
		WHERE user_id = $1
		ORDER BY created_at ASC`

	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query projects: %v", err)
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var project Project
		err := rows.Scan(
			&project.ProjectID,
			&project.UserID,
			&project.ProjectName,
			&project.Color,
			&project.IsInbox,
			&project.ParentProjectID,
			&project.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %v", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (m *ProjectModel) DeleteProjectByID(projectID uuid.UUID, userID uuid.UUID) (int64, error) {
	query := `DELETE FROM projects WHERE project_id = $1 AND user_id = $2`

	result, err := m.DB.Exec(context.Background(), query, projectID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to delete project: %v", err)
	}

	return result.RowsAffected(), nil
}
