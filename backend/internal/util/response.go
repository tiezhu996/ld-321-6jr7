package util

import (
	"errors"
	"net/http"

	"github.com/agridispatch/agridispatch/internal/constants"
	apperrors "github.com/agridispatch/agridispatch/internal/errors"
	"github.com/agridispatch/agridispatch/internal/repository"
	"github.com/gin-gonic/gin"
)

// Response 统一响应体。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: constants.CodeOK, Message: "ok", Data: data})
}

// Created 创建成功响应。
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: constants.CodeOK, Message: "ok", Data: data})
}

// Fail 业务失败响应。
func Fail(c *gin.Context, status, code int, message string) {
	c.JSON(status, Response{Code: code, Message: message})
}

// FailError 根据错误类型转换为统一响应。
func FailError(c *gin.Context, err error) {
	var validationErr *apperrors.ValidationError
	var completedErr *apperrors.TaskAlreadyCompletedError
	var notDispatchedErr *apperrors.TaskNotDispatchedError
	var repairErr *apperrors.MachineUnderRepairError
	switch {
	case errors.Is(err, repository.ErrNotFound):
		Fail(c, http.StatusNotFound, constants.CodeNotFound, err.Error())
	case errors.As(err, &validationErr):
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, validationErr.Message)
	case errors.As(err, &completedErr):
		// 重复/并发完工：第二次起以 409 冲突拒绝，不改变任何状态。
		Fail(c, http.StatusConflict, constants.CodeConflict, completedErr.Error())
	case errors.As(err, &notDispatchedErr):
		Fail(c, http.StatusConflict, constants.CodeConflict, notDispatchedErr.Error())
	case errors.As(err, &repairErr):
		Fail(c, http.StatusConflict, constants.CodeConflict, repairErr.Error())
	default:
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, err.Error())
	}
}
