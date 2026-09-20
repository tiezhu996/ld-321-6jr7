package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/agridispatch/agridispatch/internal/constants"
	apperrors "github.com/agridispatch/agridispatch/internal/errors"
	"github.com/agridispatch/agridispatch/internal/model"
	"github.com/agridispatch/agridispatch/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CompletionService 任务完工服务：登记实际作业数据、结束任务、
// 释放农机与驾驶员，并按实际工时扣减保养剩余时长。
type CompletionService struct {
	repo   *repository.DashboardRepository
	redis  *redis.Client
	logger *slog.Logger
}

func NewCompletionService(repo *repository.DashboardRepository, redis *redis.Client, logger *slog.Logger) *CompletionService {
	return &CompletionService{repo: repo, redis: redis, logger: logger}
}

// CompleteTask 完工登记。整个状态流转在单个数据库事务内完成，
// 并以条件更新（仅“已派单”可完工）保证重复/并发完工只成功一次；
// 字段缺失等校验在 handler 层拦截，任务状态不允许时这里返回冲突错误，
// 任何失败都不会留下部分改动。
func (s *CompletionService) CompleteTask(ctx context.Context, taskID string, actualHours, fuelLiters, areaMu float64) (*CompletionResultView, error) {
	task, err := s.repo.FindTask(taskID)
	if err != nil {
		return nil, err
	}
	switch task.Status {
	case constants.TaskDone:
		return nil, apperrors.NewConflict(fmt.Sprintf("任务 %s 已完工，请勿重复提交", taskID))
	case constants.TaskDispatched:
		// 唯一允许完工的状态。
	default:
		return nil, apperrors.NewConflict(fmt.Sprintf("任务 %s 当前状态为 %s，不允许完工", taskID, task.Status))
	}

	machine, err := s.repo.FindMachineByCode(task.RecommendedMachine)
	if err != nil {
		return nil, err
	}
	driver, err := s.repo.FindDriverByName(task.RecommendedDriver)
	if err != nil {
		return nil, err
	}
	reminders, err := s.repo.FindActiveRemindersByMachine(machine.Code)
	if err != nil {
		return nil, err
	}

	updatedReminders, minRemaining, maintenanceDue := applyMaintenanceDeduction(reminders, actualHours)
	fuelCost := calcFuelCost(fuelLiters)
	machineStatus := constants.MachineIdle
	machineCurrentTask := constants.MachineIdleTaskText
	if maintenanceDue {
		machineStatus = constants.MachineRepair
		machineCurrentTask = "保养到期，维修中"
	}

	record := model.WorkRecord{
		ID:          "r-" + uuid.NewString(),
		TaskID:      task.ID,
		MachineCode: machine.Code,
		DriverName:  driver.Name,
		WorkDate:    time.Now().Format("2006-01-02"),
		TaskType:    task.Type,
		ActualHours: actualHours,
		FuelLiters:  fuelLiters,
		AreaMu:      areaMu,
		FuelCost:    fuelCost,
	}

	plan := repository.CompletionPlan{
		TaskID:             task.ID,
		MachineCode:        machine.Code,
		MachineStatus:      machineStatus,
		MachineCurrentTask: machineCurrentTask,
		ActualHours:        actualHours,
		DriverName:         driver.Name,
		Record:             record,
		Reminders:          updatedReminders,
	}
	if err := s.repo.ApplyCompletion(&plan); err != nil {
		if errors.Is(err, repository.ErrAlreadyCompleted) {
			return nil, apperrors.NewConflict(fmt.Sprintf("任务 %s 已被完工，重复提交无效", taskID))
		}
		return nil, err
	}

	s.invalidate(ctx)
	s.logger.Info("task completed",
		"taskId", task.ID, "machine", machine.Code, "driver", driver.Name,
		"hours", actualHours, "maintenanceDue", maintenanceDue, "remainingHours", minRemaining)

	message := "作业记录已登记，任务结束，农机与驾驶员已恢复可用"
	if maintenanceDue {
		message = "作业记录已登记，保养剩余时长已扣减至零，农机转为维修中并暂停派单"
	}
	return &CompletionResultView{
		TaskID:         task.ID,
		Status:         constants.TaskDone,
		RecordID:       record.ID,
		MachineCode:    machine.Code,
		MachineStatus:  machineStatus,
		RemainingHours: minRemaining,
		DriverName:     driver.Name,
		DriverStatus:   constants.DriverAvailable,
		ActualHours:    actualHours,
		FuelLiters:     fuelLiters,
		AreaMu:         areaMu,
		FuelCost:       fuelCost,
		MaintenanceDue: maintenanceDue,
		Message:        message,
	}, nil
}

// invalidate 完工后使看板缓存失效，任务列表/统计/保养提醒刷新后即同步。
func (s *CompletionService) invalidate(ctx context.Context) {
	if err := s.redis.Del(ctx, constants.OverviewCacheKey).Err(); err != nil {
		s.logger.Warn("invalidate overview cache failed", "err", err)
	}
}

// CompletionResultView 完工结果视图（与 dto.CompleteTaskResult 对应）。
type CompletionResultView = completionResult

// completionResult 独立类型，避免 service 反向依赖 dto 包。
type completionResult struct {
	TaskID         string  `json:"taskId"`
	Status         string  `json:"status"`
	RecordID       string  `json:"recordId"`
	MachineCode    string  `json:"machineCode"`
	MachineStatus  string  `json:"machineStatus"`
	RemainingHours float64 `json:"remainingHours"`
	DriverName     string  `json:"driverName"`
	DriverStatus   string  `json:"driverStatus"`
	ActualHours    float64 `json:"actualHours"`
	FuelLiters     float64 `json:"fuelLiters"`
	AreaMu         float64 `json:"areaMu"`
	FuelCost       float64 `json:"fuelCost"`
	MaintenanceDue bool    `json:"maintenanceDue"`
	Message        string  `json:"message"`
}

// applyMaintenanceDeduction 按实际工时扣减该农机每条保养提醒的剩余时长，
// 扣至零的提醒级别升级为 danger；返回扣减后的提醒、最小剩余时长、
// 以及是否有提醒归零（归零则农机应转为维修中）。
func applyMaintenanceDeduction(reminders []model.MaintenanceReminder, actualHours float64) ([]model.MaintenanceReminder, float64, bool) {
	updated := make([]model.MaintenanceReminder, len(reminders))
	minRemaining := math.MaxFloat64
	due := false
	for i, r := range reminders {
		r.RemainingHours = math.Max(constants.MaintenanceRemainingZero, r.RemainingHours-actualHours)
		if r.RemainingHours == constants.MaintenanceRemainingZero {
			r.Level = constants.MaintenanceLevelDanger
			due = true
		}
		if r.RemainingHours < minRemaining {
			minRemaining = r.RemainingHours
		}
		updated[i] = r
	}
	if len(reminders) == 0 {
		minRemaining = 0
	}
	return updated, minRemaining, due
}

// calcFuelCost 按参考油价由油耗（升）折算油耗成本，保留两位小数。
func calcFuelCost(fuelLiters float64) float64 {
	return math.Round(fuelLiters*constants.FuelPricePerLiter*100) / 100
}
