package handler

import (
	"net/http"

	"github.com/agridispatch/agridispatch/internal/constants"
	"github.com/agridispatch/agridispatch/internal/dto"
	"github.com/agridispatch/agridispatch/internal/service"
	"github.com/agridispatch/agridispatch/internal/util"
	"github.com/gin-gonic/gin"
)

// TaskCompletionHandler 任务完工处理器。
type TaskCompletionHandler struct {
	completionSvc *service.CompletionService
}

func NewTaskCompletionHandler(completionSvc *service.CompletionService) *TaskCompletionHandler {
	return &TaskCompletionHandler{completionSvc: completionSvc}
}

// Complete 已派单任务完工：登记实际工时、油耗与作业面积，生成作业记录并结束任务。
// 字段缺失或非正数时直接拒绝（400），不改动任何记录、任务、农机或驾驶员状态。
func (h *TaskCompletionHandler) Complete(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "task id required")
		return
	}
	var req dto.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest,
			"完工登记失败：实际工时、油耗、作业面积均为必填且必须大于 0")
		return
	}

	result, err := h.completionSvc.CompleteTask(c.Request.Context(), taskID, req.ActualHours, req.FuelLiters, req.AreaMu)
	if err != nil {
		util.FailError(c, err)
		return
	}
	util.OK(c, result)
}
