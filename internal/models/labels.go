package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Label struct {
	LabelID   uuid.UUID `json:"label_id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type LabelModel struct {
	DB *pgxpool.Pool
}

func (m *LabelModel) AddLabel(userID uuid.UUID, name string) (Label, error) {
	query := `INSERT INTO labels (user_id, name) VALUES ($1, $2) RETURNING label_id, user_id, name, created_at`
	var label Label
	err := m.DB.QueryRow(context.Background(), query, userID, name).Scan(
		&label.LabelID,
		&label.UserID,
		&label.Name,
		&label.CreatedAt,
	)
	if err != nil {
		return Label{}, fmt.Errorf("unable to add label: %w", err)
	}
	return label, nil
}

func (m *LabelModel) EditLabelByID(labelID, userID uuid.UUID, name string) (Label, error) {
	query := `UPDATE labels SET name = $3 WHERE label_id = $1 AND user_id = $2 RETURNING label_id, user_id, name, created_at`
	var label Label
	err := m.DB.QueryRow(context.Background(), query, labelID, userID, name).Scan(
		&label.LabelID,
		&label.UserID,
		&label.Name,
		&label.CreatedAt,
	)
	if err != nil {
		return Label{}, fmt.Errorf("unable to edit label: %w", err)
	}
	return label, nil
}

func (m *LabelModel) GetLabelsByUserID(userID uuid.UUID) ([]Label, error) {
	query := `SELECT label_id, user_id, name, created_at FROM labels WHERE user_id = $1 ORDER BY name ASC`
	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get labels: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var label Label
		err := rows.Scan(&label.LabelID, &label.UserID, &label.Name, &label.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to scan label: %w", err)
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func (m *LabelModel) DeleteLabelByID(labelID, userID uuid.UUID) (int64, error) {
	query := `DELETE FROM labels WHERE label_id = $1 AND user_id = $2`
	cmdTag, err := m.DB.Exec(context.Background(), query, labelID, userID)
	if err != nil {
		return 0, fmt.Errorf("unable to delete label: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func ValidateLabel(label *Label, v *Validator) {
	if label.Name == "" {
		v.AddError("name", "Label name is required")
	}
	// Add more validation as needed (e.g., length, allowed characters)
}
