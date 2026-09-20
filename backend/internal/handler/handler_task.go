package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/agridispatch/agridispatch/internal/dto"
	apperrors "github.com/agridispatch/agridispatch/internal/errors"
	"github.com/agridispatch/agridispatch/internal/service"
	"github.com/agridispatch/agridispatch/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// TaskHandler 作业任务处理器（派单 / 完工登记）。
type TaskHandler struct {
	taskSvc  *service.TaskService
	validate *validator.Validate
}

func NewTaskHandler(taskSvc *service.TaskService) *TaskHandler {
	return &TaskHandler{taskSvc: taskSvc, validate: validator.New()}
}

// Dispatch 一键派单。
func (h *TaskHandler) Dispatch(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		util.Fail(c, http.StatusBadRequest, 40000, "task id required")
		return
	}
	res, err := h.taskSvc.Dispatch(c.Request.Context(), taskID)
	if err != nil {
		util.FailError(c, err)
		return
	}
	util.OK(c, res)
}

// Complete 完工登记：登记实际工时、油耗、作业面积，任务结束并释放农机与驾驶员。
func (h *TaskHandler) Complete(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		util.Fail(c, http.StatusBadRequest, 40000, "task id required")
		return
	}
	var req dto.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, describeBindError(err))
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, describeValidateError(err))
		return
	}
	result, err := h.taskSvc.CompleteTask(c.Request.Context(), taskID, &req)
	if err != nil {
		util.FailError(c, err)
		return
	}
	util.OK(c, result)
}

// describeBindError 将 JSON 绑定错误翻译为明确的字段缺失/类型错误提示。
func describeBindError(err error) string {
	var typeErr *jsonUnmarshalTypeError
	if asJSONTypeError(err, &typeErr) {
		return fmt.Sprintf("字段 %s 类型不正确，应为数值", typeErr.Field)
	}
	if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "invalid character") {
		return "请求体必须是 JSON，且包含 actualHours、fuelLiters、areaMu 字段"
	}
	return apperrors.NewValidation(err.Error()).Message
}

// describeValidateError 将 validator 规则错误翻译为中文提示。
func describeValidateError(err error) string {
	var invalid *validator.InvalidValidationError
	if asInvalidValidation(err, &invalid) {
		return "完工登记参数不合法"
	}
	var verrs validator.ValidationErrors
	if !asValidationErrors(err, &verrs) {
		return err.Error()
	}
	msgs := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		switch fe.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s 为必填字段", fieldLabel(fe.Field())))
		case "gt":
			msgs = append(msgs, fmt.Sprintf("%s 必须大于 %s", fieldLabel(fe.Field()), fe.Param()))
		case "gte":
			msgs = append(msgs, fmt.Sprintf("%s 不能小于 %s", fieldLabel(fe.Field()), fe.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s 不满足校验规则 %s", fieldLabel(fe.Field()), fe.Tag()))
		}
	}
	return strings.Join(msgs, "；")
}

func fieldLabel(field string) string {
	switch field {
	case "ActualHours":
		return "actualHours（实际工时）"
	case "FuelLiters":
		return "fuelLiters（油耗升数）"
	case "AreaMu":
		return "areaMu（作业面积）"
	default:
		return field
	}
}
