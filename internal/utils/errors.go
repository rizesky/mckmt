package utils

import "fmt"

// Common error message builders to reduce repetition

// ErrMarshal creates a "failed to marshal" error
func ErrMarshal(what string, err error) error {
	return fmt.Errorf("failed to marshal %s: %w", what, err)
}

// ErrUnmarshal creates a "failed to unmarshal" error
func ErrUnmarshal(what string, err error) error {
	return fmt.Errorf("failed to unmarshal %s: %w", what, err)
}

// ErrCreate creates a "failed to create" error
func ErrCreate(what string, err error) error {
	return fmt.Errorf("failed to create %s: %w", what, err)
}

// ErrGet creates a "failed to get" error
func ErrGet(what string, err error) error {
	return fmt.Errorf("failed to get %s: %w", what, err)
}

// ErrUpdate creates a "failed to update" error
func ErrUpdate(what string, err error) error {
	return fmt.Errorf("failed to update %s: %w", what, err)
}

// ErrParse creates a "failed to parse" error
func ErrParse(what string, err error) error {
	return fmt.Errorf("failed to parse %s: %w", what, err)
}

// ErrValidate creates a "failed to validate" error
func ErrValidate(what string, err error) error {
	return fmt.Errorf("failed to validate %s: %w", what, err)
}
