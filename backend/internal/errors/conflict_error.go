package errors

// ConflictError 状态冲突错误（如重复/并发完工、对维修中农机派单）。
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return "conflict: " + e.Message
}

// NewConflict 构造状态冲突错误。
func NewConflict(message string) *ConflictError {
	return &ConflictError{Message: message}
}
