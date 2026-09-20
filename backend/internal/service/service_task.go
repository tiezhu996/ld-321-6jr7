package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/agridispatch/agridispatch/internal/constants"
	"github.com/agridispatch/agridispatch/internal/dto"
	apperrors "github.com/agridispatch/agridispatch/internal/errors"
	"github.com/agridispatch/agridispatch/internal/repository"
	"github.com/redis/go-redis/v9"
)

// TaskService 作业任务服务（派单与完工登记）。
type TaskService struct {
	repo   *repository.TaskRepository
	redis  *redis.Client
	logger *slog.Logger
}

func NewTaskService(repo *repository.TaskRepository, redis *redis.Client, logger *slog.Logger) *TaskService {
	return &TaskService{repo: repo, redis: redis, logger: logger}
}

// Dispatch 一键派单：任务置为已派单，农机作业中，驾驶员作业中。
func (s *TaskService) Dispatch(ctx context.Context, taskID string) (map[string]interface{}, error) {
	if taskID == "" {
		return nil, apperrors.NewValidation("task id required")
	}
	machineCode, driverName, err := s.repo.Dispatch(taskID)
	if err != nil {
		return nil, err
	}
	s.invalidateOverview(ctx)
	s.logger.Info("task dispatched", "taskId", taskID, "machine", machineCode, "driver", driverName)
	return map[string]interface{}{
		"taskId":  taskID,
		"status":  constants.TaskDispatched,
		"message": "系统已按空闲度和驾驶员排班完成推荐派单",
	}, nil
}

// CompleteTask 完工登记：校验实际工时/油耗/作业面积字段，登记可追溯作业记录，
// 任务结束、农机与驾驶员恢复可用；保养剩余工时扣至零时农机转维修中。
// 重复或并发完工只允许成功一次（仓储层行锁 + task_id 唯一索引保证）。
func (s *TaskService) CompleteTask(ctx context.Context, taskID string, req *dto.CompleteTaskRequest) (*repository.TaskCompletionResult, error) {
	if err := validateCompleteRequest(taskID, req); err != nil {
		return nil, err
	}
	now := time.Now()
	plan := &repository.TaskCompletionPlan{
		TaskID:      taskID,
		ActualHours: *req.ActualHours,
		FuelLiters:  *req.FuelLiters,
		AreaMu:      *req.AreaMu,
		FuelCost:    round2(*req.FuelLiters * constants.DefaultFuelPriceYuan),
		WorkDate:    now.Format(constants.WorkDateLayout),
		Now:         now,
	}
	result, err := s.repo.CompleteTask(plan)
	if err != nil {
		return nil, err
	}
	s.invalidateOverview(ctx)
	s.logger.Info("task completed",
		"taskId", taskID,
		"recordId", result.Record.ID,
		"machine", result.MachineCode,
		"machineStatus", result.MachineStatus,
		"actualHours", plan.ActualHours,
		"remainingHours", result.RemainingHours,
		"maintenanceDue", result.MaintenanceDue)
	return result, nil
}

// validateCompleteRequest 校验完工登记入参：任务 ID 与工时/油耗/面积三个字段必填。
func validateCompleteRequest(taskID string, req *dto.CompleteTaskRequest) error {
	if taskID == "" {
		return apperrors.NewValidation("task id required")
	}
	if req == nil {
		return apperrors.NewValidation("completion payload required")
	}
	if req.ActualHours == nil {
		return apperrors.NewValidation("actualHours 为必填字段")
	}
	if req.FuelLiters == nil {
		return apperrors.NewValidation("fuelLiters 为必填字段")
	}
	if req.AreaMu == nil {
		return apperrors.NewValidation("areaMu 为必填字段")
	}
	if *req.ActualHours <= 0 {
		return apperrors.NewValidation(fmt.Sprintf("actualHours 必须大于 0，当前为 %.2f", *req.ActualHours))
	}
	if *req.FuelLiters < 0 {
		return apperrors.NewValidation(fmt.Sprintf("fuelLiters 不能为负数，当前为 %.2f", *req.FuelLiters))
	}
	if *req.AreaMu <= 0 {
		return apperrors.NewValidation(fmt.Sprintf("areaMu 必须大于 0，当前为 %.2f", *req.AreaMu))
	}
	return nil
}

// round2 保留两位小数。
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// invalidateOverview 使看板总览缓存失效，任务列表/统计/保养提醒刷新后同步。
func (s *TaskService) invalidateOverview(ctx context.Context) {
	if err := s.redis.Del(ctx, constants.OverviewCacheKey).Err(); err != nil {
		s.logger.Warn("invalidate overview cache failed", "err", err)
	}
}
