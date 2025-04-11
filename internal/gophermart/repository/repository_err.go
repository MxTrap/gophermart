package storage

import "fmt"

type RepositoryError struct {
	Operation string
	Err       error
}

func NewRepositoryError(operation string, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Err:       err,
	}
}

func (e *RepositoryError) Error() string {
	return fmt.Sprintf("RepositoryError: operation: %s, error: %s", e.Operation, e.Err.Error())
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}
