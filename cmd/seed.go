package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func SeedDemoData(ctx context.Context, tx pgx.Tx, userID string) error {
	// Parse userID as UUID to ensure it's valid
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid userID format: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	in3Days := today.Add(72 * time.Hour)

	// Create labels
	labelNames := []string{"urgent", "work", "personal"}
	labelIDs := make(map[string]uuid.UUID)

	for _, labelName := range labelNames {
		labelID := uuid.New()

		_, err := tx.Exec(ctx, `
			INSERT INTO labels (label_id, user_id, name, created_at)
			VALUES ($1, $2, $3, $4)
		`, labelID, userUUID, labelName, now)
		if err != nil {
			return fmt.Errorf("failed to create label %s: %w", labelName, err)
		}

		labelIDs[labelName] = labelID
	}

	urgentLabelID := labelIDs["urgent"]
	workLabelID := labelIDs["work"]
	personalLabelID := labelIDs["personal"]

	// Create tasks
	tasks := []struct {
		content     string
		priority    int16
		dueDate     *time.Time
		isCompleted bool
		labels      []string
		order       int
	}{
		{"Review quarterly budget report", 1, &tomorrow, false, []string{workLabelID.String()}, 0},
		{"Buy groceries for the week", 2, &today, false, []string{personalLabelID.String()}, 1},
		{"Schedule dentist appointment", 3, nil, false, []string{}, 2},
		{"Prepare presentation slides", 1, &in3Days, false, []string{workLabelID.String(), urgentLabelID.String()}, 3},
		{"Call mom", 4, nil, true, []string{}, 4},
	}

	for _, t := range tasks {
		taskID := uuid.New()
		var completedAt *time.Time
		if t.isCompleted {
			completedAt = &now
		}

		// Convert labels slice to JSONB
		labelsJSON, err := json.Marshal(t.labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels for task %s: %w", t.content, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO tasks (task_id, project_id, user_id, content, priority, due_date, is_completed, completed_at, labels, "order", created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11)
		`, taskID, nil, userUUID, t.content, t.priority, t.dueDate, t.isCompleted, completedAt, labelsJSON, t.order, now)
		if err != nil {
			return fmt.Errorf("failed to create task %s: %w", t.content, err)
		}
	}

	return nil
}
