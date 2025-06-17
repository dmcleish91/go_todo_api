package models

// Validator provides a simple way to collect validation errors.
type Validator struct {
	Errors map[string]string
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// AddError adds an error message for a given field if it doesn't already exist.
func (v *Validator) AddError(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

// Check adds an error message if the condition is false.
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

// Valid returns true if there are no errors.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// ValidateTask validates a Task object
func ValidateTask(task *Task, v *Validator) {
	// Validate due date format if provided
	if task.DueDate != nil {
		// Check if the date matches MM/DD/YYYY format
		dateStr := task.DueDate.Format("01/02/2006")
		v.Check(dateStr != "", "due_date", "Due date must be in MM/DD/YYYY format")
	}

	// Validate date_datetime format if provided
	if task.DateDatetime != nil {
		// Check if the time matches XX:XX AM/PM format
		timeStr := task.DateDatetime.Format("03:04 PM")
		v.Check(timeStr != "", "date_datetime", "Date time must be in XX:XX AM/PM format")
	}
}
