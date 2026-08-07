package apperror

import "fmt"

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("retryable error: %v", e.Err)
	}
	return "retryable error"
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}
