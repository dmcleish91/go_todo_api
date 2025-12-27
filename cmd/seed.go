package main

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func SeedDemoData(ctx context.Context, userID string) error {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	in3Days := today.Add(72 * time.Hour)

	// Create Inbox project
	inboxID := uuid.New().String()
	_, err := db.Exec(ctx, `
		INSERT INTO projects (project_id, user_id, project_name, color, is_inbox, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, inboxID, userID, "Inbox", nil, true, now)
	if err != nil {
		return err
	}

	// Create labels
	urgentLabelID := uuid.New().String()
	workLabelID := uuid.New().String()
	personalLabelID := uuid.New().String()

	labels := []struct {
		id, name string
	}{
		{urgentLabelID, "urgent"},
		{workLabelID, "work"},
		{personalLabelID, "personal"},
	}

	for _, l := range labels {
		_, err := db.Exec(ctx, `
			INSERT INTO labels (label_id, user_id, name, created_at)
			VALUES ($1, $2, $3, $4)
		`, l.id, userID, l.name, now)
		if err != nil {
			return err
		}
	}

	// Create tasks
	tasks := []struct {
		content     string
		priority    int
		dueDate     *time.Time
		isCompleted bool
		labels      string // JSON array
		order       int
	}{
		{"Review quarterly budget report", 1, &tomorrow, false, `["` + workLabelID + `"]`, 0},
		{"Buy groceries for the week", 2, &today, false, `["` + personalLabelID + `"]`, 1},
		{"Schedule dentist appointment", 3, nil, false, `[]`, 2},
		{"Prepare presentation slides", 1, &in3Days, false, `["` + workLabelID + `","` + urgentLabelID + `"]`, 3},
		{"Call mom", 4, nil, true, `[]`, 4},
	}

	for _, t := range tasks {
		taskID := uuid.New().String()
		var completedAt *time.Time
		if t.isCompleted {
			completedAt = &now
		}

		_, err := db.Exec(ctx, `
			INSERT INTO tasks (task_id, project_id, user_id, content, priority, due_date, is_completed, completed_at, labels, "order", created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, taskID, inboxID, userID, t.content, t.priority, t.dueDate, t.isCompleted, completedAt, t.labels, t.order, now)
		if err != nil {
			return err
		}
	}

	return nil
}

