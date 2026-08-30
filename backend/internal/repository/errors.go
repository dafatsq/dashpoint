package repository

// InternalError marks repository failures whose details (driver messages,
// SQLSTATE codes, transaction state) must never reach API clients. Handlers
// check for it with errors.As and substitute a generic message while logging
// the original error server-side.
type InternalError struct {
	Err error
}

func (e *InternalError) Error() string { return e.Err.Error() }
func (e *InternalError) Unwrap() error { return e.Err }

// InternalError wraps an infrastructure failure (transactions, inserts,
// driver errors) so handlers can distinguish it from domain-validation
// messages, which are safe to show to users verbatim.
func NewInternalError(err error) error {
	return &InternalError{Err: err}
}
