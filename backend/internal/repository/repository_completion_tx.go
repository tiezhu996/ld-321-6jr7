package repository

import (
	"errors"
	"fmt"

	"github.com/agridispatch/agridispatch/internal/constants"
	"github.com/agridispatch/agridispatch/internal/model"
	"gorm.io/gorm"
)

// ErrAlreadyCompleted 任务已被并发/重复完工（条件更新未命中任何行）。
var ErrAlreadyCompleted = errors.New("task already completed")

// CompletionPlan 完工状态变更计划，由 service 层按业务规则计算，
// repository 在单个事务内原子落库。
type CompletionPlan struct {
	TaskID             string
	MachineCode        string
	MachineStatus      string // 恢复空闲，或保养归零转为维修中
	MachineCurrentTask string
	ActualHours        float64
	DriverName         string
	Record             model.WorkRecord
	Reminders          []model.MaintenanceReminder
}

// ApplyCompletion 在事务内完成全部完工改动：
//  1. 条件更新抢占完工权（仅“已派单”任务可改为“已完成”，受行锁保护）；
//  2. 插入可追溯作业记录；
//  3. 农机恢复可用（或转维修中）、累计工时、清空当前任务；
//  4. 驾驶员恢复可派单；
//  5. 逐条扣减保养提醒剩余时长。
//
// 任一步骤失败整体回滚；并发完工时只有第一个事务的条件更新命中 1 行，
// 其余事务返回 ErrAlreadyCompleted，不会产生任何记录或状态改动。
func (r *DashboardRepository) ApplyCompletion(plan *CompletionPlan) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.FarmTask{}).
			Where("id = ? AND status = ?", plan.TaskID, constants.TaskDispatched).
			Update("status", constants.TaskDone)
		if res.Error != nil {
			return fmt.Errorf("mark task done: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return ErrAlreadyCompleted
		}

		if err := tx.Create(&plan.Record).Error; err != nil {
			return fmt.Errorf("create work record: %w", err)
		}

		if err := tx.Model(&model.Machine{}).
			Where("code = ?", plan.MachineCode).
			Updates(map[string]interface{}{
				"status":       plan.MachineStatus,
				"current_task": plan.MachineCurrentTask,
				"work_hours":   gorm.Expr("work_hours + ?", plan.ActualHours),
			}).Error; err != nil {
			return fmt.Errorf("release machine: %w", err)
		}

		if err := tx.Model(&model.Driver{}).
			Where("name = ?", plan.DriverName).
			Update("status", constants.DriverAvailable).Error; err != nil {
			return fmt.Errorf("release driver: %w", err)
		}

		for i := range plan.Reminders {
			if err := tx.Model(&model.MaintenanceReminder{}).
				Where("id = ?", plan.Reminders[i].ID).
				Updates(map[string]interface{}{
					"remaining_hours": plan.Reminders[i].RemainingHours,
					"level":           plan.Reminders[i].Level,
				}).Error; err != nil {
				return fmt.Errorf("deduct reminder: %w", err)
			}
		}
		return nil
	})
}
