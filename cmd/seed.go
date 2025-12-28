package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func SeedDemoData(ctx context.Context, userID string) error {
	// Parse userID as UUID to ensure it's valid
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid userID format: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	in3Days := today.Add(72 * time.Hour)

	// Create Inbox project (or get existing one if it already exists due to race condition)
	var inboxID uuid.UUID
	err = db.QueryRow(ctx, `
		SELECT project_id FROM projects WHERE user_id = $1 AND is_inbox = true LIMIT 1
	`, userUUID).Scan(&inboxID)

	if err != nil {
		// No inbox project exists, create one
		inboxID = uuid.New()
		_, err = db.Exec(ctx, `
			INSERT INTO projects (project_id, user_id, project_name, color, is_inbox, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, inboxID, userUUID, "Inbox", nil, true, now)
		if err != nil {
			// If insert fails, try to get existing inbox again (race condition)
			err = db.QueryRow(ctx, `
				SELECT project_id FROM projects WHERE user_id = $1 AND is_inbox = true LIMIT 1
			`, userUUID).Scan(&inboxID)
			if err != nil {
				return fmt.Errorf("failed to create or get inbox project: %w", err)
			}
		}
	}

	// Create labels (or get existing ones if they already exist due to race condition)
	labelNames := []string{"urgent", "work", "personal"}
	labelIDs := make(map[string]uuid.UUID)

	for _, labelName := range labelNames {
		labelID := uuid.New()

		// Try to insert, but if it already exists, that's fine
		_, err := db.Exec(ctx, `
			INSERT INTO labels (label_id, user_id, name, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, name) DO NOTHING
		`, labelID, userUUID, labelName, now)
		if err != nil {
			return fmt.Errorf("failed to create label %s: %w", labelName, err)
		}

		// Get the actual label ID (either the one we just created or the existing one)
		var actualLabelID uuid.UUID
		err = db.QueryRow(ctx, `
			SELECT label_id FROM labels WHERE user_id = $1 AND name = $2
		`, userUUID, labelName).Scan(&actualLabelID)
		if err != nil {
			return fmt.Errorf("failed to get label ID for %s: %w", labelName, err)
		}

		labelIDs[labelName] = actualLabelID
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

		_, err = db.Exec(ctx, `
			INSERT INTO tasks (task_id, project_id, user_id, content, priority, due_date, is_completed, completed_at, labels, "order", created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11)
		`, taskID, inboxID, userUUID, t.content, t.priority, t.dueDate, t.isCompleted, completedAt, labelsJSON, t.order, now)
		if err != nil {
			return fmt.Errorf("failed to create task %s: %w", t.content, err)
		}
	}

	return nil
}
