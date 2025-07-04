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
