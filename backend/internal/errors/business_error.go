package errors

import "fmt"

// BusinessError 业务错误。
type BusinessError struct {
	Code    int
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// New 构造业务错误。
func New(code int, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}

// ValidationError 参数校验错误。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "validation: " + e.Message
}

// NewValidation 构造参数校验错误。
func NewValidation(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// MachineOfflineError 农机离线错误。
type MachineOfflineError struct {
	MachineCode string
}

func (e *MachineOfflineError) Error() string {
	return fmt.Sprintf("machine %s is offline", e.MachineCode)
}

// TaskNotDispatchedError 任务未处于已派单状态，不允许完工登记。
type TaskNotDispatchedError struct {
	TaskID string
	Status string
}

func (e *TaskNotDispatchedError) Error() string {
	return fmt.Sprintf("task %s status is %s, only dispatched tasks can be completed", e.TaskID, e.Status)
}

// TaskAlreadyCompletedError 任务已完工，重复/并发完工只允许成功一次。
type TaskAlreadyCompletedError struct {
	TaskID string
}

func (e *TaskAlreadyCompletedError) Error() string {
	return fmt.Sprintf("task %s is already completed", e.TaskID)
}

// MachineUnderRepairError 农机维修中，阻止继续派单。
type MachineUnderRepairError struct {
	MachineCode string
}

func (e *MachineUnderRepairError) Error() string {
	return fmt.Sprintf("machine %s is under repair and cannot be dispatched", e.MachineCode)
}

// CompleteFieldsError 完工登记字段缺失错误。
type CompleteFieldsError struct {
	Field string
}

func (e *CompleteFieldsError) Error() string {
	return fmt.Sprintf("completion field %s is required", e.Field)
}
